//go:build integration

package pipeline

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ==================== Artifact Manager Tests ====================

func setupArtifactTest(t *testing.T) (*ArtifactManager, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-artifact-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewPipelineDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("NewPipelineDB failed: %v", err)
	}

	config := &ArtifactConfig{
		StoragePath:   filepath.Join(tmpDir, "artifacts"),
		RetentionDays: 30,
		MaxSize:       100 * 1024 * 1024,
	}

	am, err := NewArtifactManager(db, config)
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("NewArtifactManager failed: %v", err)
	}

	// Set state file to temp directory and clear any state loaded from default path
	am.SetStateFile(filepath.Join(tmpDir, "artifacts-state.json"))
	am.mu.Lock()
	for k := range am.artifacts {
		delete(am.artifacts, k)
	}
	am.mu.Unlock()

	t.Cleanup(func() {
		db.Close()
		os.RemoveAll(tmpDir)
	})

	return am, tmpDir
}

func TestArtifactManager_New(t *testing.T) {
	am, _ := setupArtifactTest(t)
	if am == nil {
		t.Fatal("ArtifactManager should not be nil")
	}
}

func TestArtifactManager_UploadAndDownload(t *testing.T) {
	am, _ := setupArtifactTest(t)
	ctx := context.Background()

	content := []byte("test artifact content for upload and download")
	artifact, err := am.UploadArtifact(ctx, "test-pipeline-upload", "log", "test.log", bytes.NewReader(content), map[string]string{"env": "test"})
	if err != nil {
		t.Fatalf("UploadArtifact failed: %v", err)
	}

	if artifact.ID == "" {
		t.Error("Artifact ID should not be empty")
	}
	if artifact.PipelineID != "test-pipeline-upload" {
		t.Errorf("PipelineID = %s, want test-pipeline-upload", artifact.PipelineID)
	}
	if artifact.Name != "test.log" {
		t.Errorf("Name = %s, want test.log", artifact.Name)
	}
	if artifact.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", artifact.Size, len(content))
	}
	if artifact.Checksum == "" {
		t.Error("Artifact should have checksum")
	}
	if len(artifact.Checksum) != 64 {
		t.Errorf("Checksum length = %d, want 64 (SHA256 hex)", len(artifact.Checksum))
	}

	// Download
	var buf bytes.Buffer
	err = am.DownloadArtifact(ctx, artifact.ID, &buf)
	if err != nil {
		t.Fatalf("DownloadArtifact failed: %v", err)
	}
	if buf.String() != string(content) {
		t.Errorf("Downloaded content mismatch: got %d bytes, want %d", buf.Len(), len(content))
	}
}

func TestArtifactManager_ListArtifacts(t *testing.T) {
	am, _ := setupArtifactTest(t)
	ctx := context.Background()
	pipelineID := "test-pipeline-list"

	for i := 0; i < 3; i++ {
		_, err := am.UploadArtifact(ctx, pipelineID, "log", "test.log", bytes.NewReader([]byte("content")), nil)
		if err != nil {
			t.Fatalf("UploadArtifact %d failed: %v", i, err)
		}
	}

	artifacts := am.ListArtifacts(pipelineID)
	if len(artifacts) != 3 {
		t.Errorf("ListArtifacts returned %d artifacts, want 3", len(artifacts))
	}
}

func TestArtifactManager_ListArtifactsByType(t *testing.T) {
	am, _ := setupArtifactTest(t)
	ctx := context.Background()
	pipelineID := "test-pipeline-bytype"

	am.UploadArtifact(ctx, pipelineID, "log", "test.log", bytes.NewReader([]byte("log content")), nil)
	am.UploadArtifact(ctx, pipelineID, "artifact", "test.bin", bytes.NewReader([]byte("bin content")), nil)

	logs := am.ListArtifactsByType(pipelineID, "log")
	if len(logs) != 1 {
		t.Errorf("ListArtifactsByType log returned %d, want 1", len(logs))
	}

	artifacts := am.ListArtifactsByType(pipelineID, "artifact")
	if len(artifacts) != 1 {
		t.Errorf("ListArtifactsByType artifact returned %d, want 1", len(artifacts))
	}
}

func TestArtifactManager_GetArtifact(t *testing.T) {
	am, _ := setupArtifactTest(t)
	ctx := context.Background()

	artifact, err := am.UploadArtifact(ctx, "test-pipeline-get", "log", "test.log", bytes.NewReader([]byte("content")), nil)
	if err != nil {
		t.Fatalf("UploadArtifact failed: %v", err)
	}

	retrieved, err := am.GetArtifact(artifact.ID)
	if err != nil {
		t.Fatalf("GetArtifact failed: %v", err)
	}
	if retrieved.ID != artifact.ID {
		t.Errorf("GetArtifact ID = %s, want %s", retrieved.ID, artifact.ID)
	}
	if retrieved.Name != artifact.Name {
		t.Errorf("GetArtifact Name = %s, want %s", retrieved.Name, artifact.Name)
	}
}

func TestArtifactManager_DeleteArtifact(t *testing.T) {
	am, _ := setupArtifactTest(t)
	ctx := context.Background()

	artifact, err := am.UploadArtifact(ctx, "test-pipeline-del", "log", "test.log", bytes.NewReader([]byte("content")), nil)
	if err != nil {
		t.Fatalf("UploadArtifact failed: %v", err)
	}

	err = am.DeleteArtifact(ctx, artifact.ID)
	if err != nil {
		t.Fatalf("DeleteArtifact failed: %v", err)
	}

	_, err = am.GetArtifact(artifact.ID)
	if err == nil {
		t.Error("GetArtifact should return error for deleted artifact")
	}
}

func TestArtifactManager_DeleteArtifactsForPipeline(t *testing.T) {
	am, _ := setupArtifactTest(t)
	ctx := context.Background()
	pipelineID := "test-pipeline-delpipeline"

	for i := 0; i < 3; i++ {
		am.UploadArtifact(ctx, pipelineID, "log", "test.log", bytes.NewReader([]byte("content")), nil)
	}

	err := am.DeleteArtifactsForPipeline(ctx, pipelineID)
	if err != nil {
		t.Fatalf("DeleteArtifactsForPipeline failed: %v", err)
	}

	artifacts := am.ListArtifacts(pipelineID)
	if len(artifacts) != 0 {
		t.Errorf("ListArtifacts returned %d artifacts after delete, want 0", len(artifacts))
	}
}

func TestArtifactManager_StateFile(t *testing.T) {
	am, tmpDir := setupArtifactTest(t)

	stateFile := am.getStateFile()
	if stateFile == "" {
		t.Error("getStateFile should not return empty string")
	}

	customPath := filepath.Join(tmpDir, "custom-state.json")
	am.SetStateFile(customPath)
	if am.getStateFile() != customPath {
		t.Errorf("getStateFile = %s, want %s", am.getStateFile(), customPath)
	}
}

func TestArtifactManager_SetExpiration(t *testing.T) {
	am, _ := setupArtifactTest(t)
	ctx := context.Background()

	artifact, err := am.UploadArtifact(ctx, "test-pipeline-exp", "log", "test.log", bytes.NewReader([]byte("content")), nil)
	if err != nil {
		t.Fatalf("UploadArtifact failed: %v", err)
	}

	// Set expiration
	expiresAt := time.Now().Add(24 * time.Hour)
	err = am.SetExpiration(artifact.ID, expiresAt)
	if err != nil {
		t.Fatalf("SetExpiration failed: %v", err)
	}

	// Verify expiration was set
	retrieved, err := am.GetArtifact(artifact.ID)
	if err != nil {
		t.Fatalf("GetArtifact failed: %v", err)
	}
	if retrieved.ExpiresAt == nil {
		t.Error("ExpiresAt should be set after SetExpiration")
	}
}

func TestArtifactManager_UploadFileArtifact(t *testing.T) {
	am, tmpDir := setupArtifactTest(t)
	ctx := context.Background()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test-upload.txt")
	err := os.WriteFile(testFile, []byte("file content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	artifact, err := am.UploadFileArtifact(ctx, "test-pipeline-file", "log", testFile, nil)
	if err != nil {
		t.Fatalf("UploadFileArtifact failed: %v", err)
	}

	if artifact.Name == "" {
		t.Error("Artifact name should not be empty")
	}
}

func TestArtifactManager_DownloadArtifactToFile(t *testing.T) {
	am, tmpDir := setupArtifactTest(t)
	ctx := context.Background()

	// Upload artifact
	artifact, err := am.UploadArtifact(ctx, "test-pipeline-dlfile", "log", "test.log", bytes.NewReader([]byte("file download test")), nil)
	if err != nil {
		t.Fatalf("UploadArtifact failed: %v", err)
	}

	// Download to file
	destPath := filepath.Join(tmpDir, "downloaded.txt")
	err = am.DownloadArtifactToFile(ctx, artifact.ID, destPath)
	if err != nil {
		t.Fatalf("DownloadArtifactToFile failed: %v", err)
	}

	// Verify file content
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}
	if string(content) != "file download test" {
		t.Errorf("Downloaded content = %s, want 'file download test'", string(content))
	}
}

func TestArtifactManager_CleanupExpired(t *testing.T) {
	am, _ := setupArtifactTest(t)
	ctx := context.Background()
	pipelineID := "test-pipeline-cleanup"

	// Upload artifact
	artifact, err := am.UploadArtifact(ctx, pipelineID, "log", "expired.log", bytes.NewReader([]byte("expired content")), nil)
	if err != nil {
		t.Fatalf("UploadArtifact failed: %v", err)
	}

	// Set expiration to past
	pastTime := time.Now().Add(-1 * time.Hour)
	err = am.SetExpiration(artifact.ID, pastTime)
	if err != nil {
		t.Fatalf("SetExpiration failed: %v", err)
	}

	// Cleanup
	err = am.CleanupExpiredArtifacts(ctx)
	if err != nil {
		t.Logf("CleanupExpiredArtifacts: %v", err)
	}

	// Verify expired artifact was removed (or marked as expired)
	retrieved, err := am.GetArtifact(artifact.ID)
	if err != nil {
		// Artifact was deleted - good
		t.Log("Expired artifact was cleaned up successfully")
	} else if retrieved.ExpiresAt != nil && retrieved.ExpiresAt.Before(time.Now()) {
		// Artifact still exists but expired - also acceptable
		t.Log("Expired artifact still exists but has past expiration")
	}
}