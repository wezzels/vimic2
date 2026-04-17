// Package pipeline provides artifact management
package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ArtifactManager manages pipeline artifacts
type ArtifactManager struct {
	db          *PipelineDB
	storagePath string
	stateFile   string
	artifacts   map[string]*Artifact
	mu          sync.RWMutex
}

// Artifact represents a pipeline artifact
type Artifact struct {
	ID          string            `json:"id"`
	PipelineID  string            `json:"pipeline_id"`
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Size        int64             `json:"size"`
	Checksum    string            `json:"checksum"`
	ContentType string            `json:"content_type"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// ArtifactConfig represents artifact manager configuration
type ArtifactConfig struct {
	StoragePath   string `json:"storage_path"`
	RetentionDays int    `json:"retention_days"`
	MaxSize       int64  `json:"max_size"`
}

// NewArtifactManager creates a new artifact manager
func NewArtifactManager(db *PipelineDB, config *ArtifactConfig) (*ArtifactManager, error) {
	if config == nil {
		config = &ArtifactConfig{
			StoragePath:   filepath.Join(os.Getenv("HOME"), ".vimic2", "artifacts"),
			RetentionDays: 30,
			MaxSize:       100 * 1024 * 1024, // 100 MB
		}
	}

	am := &ArtifactManager{
		db:          db,
		storagePath: config.StoragePath,
		artifacts:   make(map[string]*Artifact),
	}

	// Create storage directory
	if err := os.MkdirAll(am.storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Load state
	if err := am.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return am, nil
}

// loadState loads artifact state from disk
func (am *ArtifactManager) loadState() error {
	stateFile := am.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
		}
		return err
	}

	var artifacts []*Artifact
	if err := json.Unmarshal(data, &artifacts); err != nil {
		return err
	}

	for _, artifact := range artifacts {
		am.artifacts[artifact.ID] = artifact
	}

	return nil
}

// saveState saves artifact state to disk
// NOTE: Caller must hold am.mu (write) lock before calling this.
func (am *ArtifactManager) saveState() error {
	artifacts := make([]*Artifact, 0, len(am.artifacts))
	for _, artifact := range am.artifacts {
		artifacts = append(artifacts, artifact)
	}

	data, err := json.MarshalIndent(artifacts, "", "  ")
	if err != nil {
		return err
	}

	stateFile := am.getStateFile()
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, data, 0644)
}

// getStateFile returns the state file path
func (am *ArtifactManager) getStateFile() string {
	if am.stateFile != "" {
		return am.stateFile
	}
	return filepath.Join(os.Getenv("HOME"), ".vimic2", "artifacts-state.json")
}

// SetStateFile sets the state file path
func (am *ArtifactManager) SetStateFile(path string) {
	am.stateFile = path
}

// UploadArtifact uploads an artifact
func (am *ArtifactManager) UploadArtifact(ctx context.Context, pipelineID, artifactType, name string, reader io.Reader, metadata map[string]string) (*Artifact, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Generate artifact ID
	artifactID := generateArtifactID()

	// Create artifact directory
	artifactDir := filepath.Join(am.storagePath, pipelineID, artifactType)
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create artifact directory: %w", err)
	}

	// Create artifact file
	artifactPath := filepath.Join(artifactDir, name)
	file, err := os.Create(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create artifact file: %w", err)
	}
	defer file.Close()

	// Copy content and calculate checksum
	hash := sha256.New()
	writer := io.MultiWriter(file, hash)

	size, err := io.Copy(writer, reader)
	if err != nil {
		os.Remove(artifactPath)
		return nil, fmt.Errorf("failed to write artifact: %w", err)
	}

	checksum := hex.EncodeToString(hash.Sum(nil))

	// Create artifact
	artifact := &Artifact{
		ID:         artifactID,
		PipelineID: pipelineID,
		Type:       artifactType,
		Name:       name,
		Path:       artifactPath,
		Size:       size,
		Checksum:   checksum,
		Metadata:   metadata,
		CreatedAt:  time.Now(),
	}

	am.artifacts[artifactID] = artifact

	// Save to database
	dbArtifact := &Artifact{
		ID:         artifact.ID,
		PipelineID: artifact.PipelineID,
		Type:       artifact.Type,
		Name:       artifact.Name,
		Path:       artifact.Path,
		Size:       artifact.Size,
		Checksum:   artifact.Checksum,
		CreatedAt:  artifact.CreatedAt,
	}
	if err := am.db.SaveArtifact(ctx, dbArtifact); err != nil {
		os.Remove(artifactPath)
		delete(am.artifacts, artifactID)
		return nil, fmt.Errorf("failed to save artifact: %w", err)
	}

	// Save state
	if err := am.saveState(); err != nil {
		am.db.DeleteArtifact(ctx, artifactID)
		os.Remove(artifactPath)
		delete(am.artifacts, artifactID)
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	return artifact, nil
}

// UploadFileArtifact uploads a file as an artifact
func (am *ArtifactManager) UploadFileArtifact(ctx context.Context, pipelineID, artifactType, filePath string, metadata map[string]string) (*Artifact, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	name := filepath.Base(filePath)
	return am.UploadArtifact(ctx, pipelineID, artifactType, name, file, metadata)
}

// DownloadArtifact downloads an artifact
func (am *ArtifactManager) DownloadArtifact(ctx context.Context, artifactID string, writer io.Writer) error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	artifact, ok := am.artifacts[artifactID]
	if !ok {
		return fmt.Errorf("artifact not found: %s", artifactID)
	}

	file, err := os.Open(artifact.Path)
	if err != nil {
		return fmt.Errorf("failed to open artifact: %w", err)
	}
	defer file.Close()

	// Verify checksum
	hash := sha256.New()
	writerCopy := io.MultiWriter(writer, hash)

	if _, err := io.Copy(writerCopy, file); err != nil {
		return fmt.Errorf("failed to read artifact: %w", err)
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	if checksum != artifact.Checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", artifact.Checksum, checksum)
	}

	return nil
}

// DownloadArtifactToFile downloads an artifact to a file
func (am *ArtifactManager) DownloadArtifactToFile(ctx context.Context, artifactID, destPath string) error {
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return am.DownloadArtifact(ctx, artifactID, file)
}

// GetArtifact returns an artifact by ID
func (am *ArtifactManager) GetArtifact(artifactID string) (*Artifact, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	artifact, ok := am.artifacts[artifactID]
	if !ok {
		return nil, fmt.Errorf("artifact not found: %s", artifactID)
	}

	return artifact, nil
}

// ListArtifacts lists artifacts for a pipeline
func (am *ArtifactManager) ListArtifacts(pipelineID string) []*Artifact {
	am.mu.RLock()
	defer am.mu.RUnlock()

	artifacts := make([]*Artifact, 0)
	for _, artifact := range am.artifacts {
		if artifact.PipelineID == pipelineID {
			artifacts = append(artifacts, artifact)
		}
	}

	return artifacts
}

// ListArtifactsByType lists artifacts by type
func (am *ArtifactManager) ListArtifactsByType(pipelineID, artifactType string) []*Artifact {
	am.mu.RLock()
	defer am.mu.RUnlock()

	artifacts := make([]*Artifact, 0)
	for _, artifact := range am.artifacts {
		if artifact.PipelineID == pipelineID && artifact.Type == artifactType {
			artifacts = append(artifacts, artifact)
		}
	}

	return artifacts
}

// DeleteArtifact deletes an artifact
func (am *ArtifactManager) DeleteArtifact(ctx context.Context, artifactID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	artifact, ok := am.artifacts[artifactID]
	if !ok {
		return fmt.Errorf("artifact not found: %s", artifactID)
	}

	// Delete file
	if err := os.Remove(artifact.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete artifact file: %w", err)
	}

	// Delete from database
	if err := am.db.DeleteArtifact(ctx, artifactID); err != nil {
		return fmt.Errorf("failed to delete artifact from database: %w", err)
	}

	// Delete from memory
	delete(am.artifacts, artifactID)

	// Save state
	return am.saveState()
}

// DeleteArtifactsForPipeline deletes all artifacts for a pipeline
func (am *ArtifactManager) DeleteArtifactsForPipeline(ctx context.Context, pipelineID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for _, artifact := range am.artifacts {
		if artifact.PipelineID == pipelineID {
			os.Remove(artifact.Path)
			am.db.DeleteArtifact(ctx, artifact.ID)
			delete(am.artifacts, artifact.ID)
		}
	}

	// Remove directory
	dir := filepath.Join(am.storagePath, pipelineID)
	os.RemoveAll(dir)

	// Save state
	return am.saveState()
}

// CleanupExpiredArtifacts removes expired artifacts
func (am *ArtifactManager) CleanupExpiredArtifacts(ctx context.Context) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Time{}
	cleaned := 0

	for id, artifact := range am.artifacts {
		if artifact.ExpiresAt != nil && artifact.ExpiresAt.Before(now) {
			os.Remove(artifact.Path)
			am.db.DeleteArtifact(ctx, id)
			delete(am.artifacts, id)
			cleaned++
		}
	}

	// Save state
	if err := am.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	fmt.Printf("[ArtifactManager] Cleaned up %d expired artifacts\n", cleaned)
	return nil
}

// SetExpiration sets an expiration time for an artifact
func (am *ArtifactManager) SetExpiration(artifactID string, expiresAt time.Time) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	artifact, ok := am.artifacts[artifactID]
	if !ok {
		return fmt.Errorf("artifact not found: %s", artifactID)
	}

	artifact.ExpiresAt = &expiresAt

	return am.saveState()
}

// GetStats returns artifact statistics
func (am *ArtifactManager) GetStats() map[string]int {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var totalSize int64
	for _, artifact := range am.artifacts {
		totalSize += artifact.Size
	}

	return map[string]int{
		"total":      len(am.artifacts),
		"total_size": int(totalSize),
	}
}

// Helper functions

func generateArtifactID() string {
	return fmt.Sprintf("artifact-%s-%d", randomString(8), time.Now().Unix())
}
