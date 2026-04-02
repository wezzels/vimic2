// Package pool provides VM pool management with QEMU backing files
package pool

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/pipeline"
)

// TemplateManager manages QEMU backing file templates
type TemplateManager struct {
	basePath    string
	overlayPath string
	templates   map[string]*pipeline.Template
	overlays    map[string]*Overlay
	mu          sync.RWMutex
}

// Overlay represents a copy-on-write overlay image
type Overlay struct {
	ID          string    `json:"id"`
	TemplateID  string    `json:"template_id"`
	VMID        string    `json:"vm_id"`
	Path        string    `json:"path"`
	ActualSize  int64     `json:"actual_size"`
	CreatedAt   time.Time `json:"created_at"`
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(basePath, overlayPath string) (*TemplateManager, error) {
	// Create directories if they don't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}
	if err := os.MkdirAll(overlayPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create overlay path: %w", err)
	}

	tm := &TemplateManager{
		basePath:    basePath,
		overlayPath: overlayPath,
		templates:   make(map[string]*pipeline.Template),
		overlays:    make(map[string]*Overlay),
	}

	// Load existing templates
	if err := tm.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	// Load existing overlays
	if err := tm.loadOverlays(); err != nil {
		return nil, fmt.Errorf("failed to load overlays: %w", err)
	}

	return tm, nil
}

// loadTemplates loads existing templates from disk
func (tm *TemplateManager) loadTemplates() error {
	templateFile := filepath.Join(tm.basePath, "templates.json")
	data, err := ioutil.ReadFile(templateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No templates file yet
		}
		return err
	}

	var templates []*pipeline.Template
	if err := json.Unmarshal(data, &templates); err != nil {
		return err
	}

	for _, t := range templates {
		tm.templates[t.ID] = t
	}

	return nil
}

// loadOverlays loads existing overlays from disk
func (tm *TemplateManager) loadOverlays() error {
	overlayFile := filepath.Join(tm.overlayPath, "overlays.json")
	data, err := ioutil.ReadFile(overlayFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No overlays file yet
		}
		return err
	}

	var overlays []*Overlay
	if err := json.Unmarshal(data, &overlays); err != nil {
		return err
	}

	for _, o := range overlays {
		// Skip destroyed overlays
		if o.DestroyedAt != nil {
			continue
		}
		tm.overlays[o.ID] = o
	}

	return nil
}

// saveTemplates saves templates to disk
func (tm *TemplateManager) saveTemplates() error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	templates := make([]*pipeline.Template, 0, len(tm.templates))
	for _, t := range tm.templates {
		templates = append(templates, t)
	}

	data, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		return err
	}

	templateFile := filepath.Join(tm.basePath, "templates.json")
	return ioutil.WriteFile(templateFile, data, 0644)
}

// saveOverlays saves overlays to disk
func (tm *TemplateManager) saveOverlays() error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	overlays := make([]*Overlay, 0, len(tm.overlays))
	for _, o := range tm.overlays {
		overlays = append(overlays, o)
	}

	data, err := json.MarshalIndent(overlays, "", "  ")
	if err != nil {
		return err
	}

	overlayFile := filepath.Join(tm.overlayPath, "overlays.json")
	return ioutil.WriteFile(overlayFile, data, 0644)
}

// CreateTemplate creates a new base template
func (tm *TemplateManager) CreateTemplate(name, baseImage string, size int64, packages []string) (*pipeline.Template, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if template already exists
	for _, t := range tm.templates {
		if t.Name == name {
			return nil, fmt.Errorf("template already exists: %s", name)
		}
	}

	templatePath := filepath.Join(tm.basePath, name+".qcow2")

	// Create base image
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", templatePath, fmt.Sprintf("%d", size))
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create template: %w: %s", err, output)
	}

	// Calculate checksum
	checksum, err := tm.calculateChecksum(templatePath)
	if err != nil {
		os.Remove(templatePath)
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Make read-only
	if err := os.Chmod(templatePath, 0444); err != nil {
		os.Remove(templatePath)
		return nil, fmt.Errorf("failed to make template read-only: %w", err)
	}

	template := &pipeline.Template{
		ID:        generateID("tpl"),
		Name:      name,
		BaseImage: templatePath,
		Size:      size,
		Packages:  packages,
		Checksum:  checksum,
		ReadOnly:  true,
		CreatedAt: time.Now(),
	}

	tm.templates[template.ID] = template
	if err := tm.saveTemplates(); err != nil {
		delete(tm.templates, template.ID)
		os.Remove(templatePath)
		return nil, err
	}

	return template, nil
}

// CreateOverlay creates a copy-on-write overlay from a base template
func (tm *TemplateManager) CreateOverlay(templateID, vmID string) (*Overlay, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	template, ok := tm.templates[templateID]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	overlayPath := filepath.Join(tm.overlayPath, fmt.Sprintf("%s-overlay.qcow2", vmID))

	// Create overlay: qemu-img create -f qcow2 -F qcow2 -b base.qcow2 overlay.qcow2
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2",
		"-F", "qcow2", "-b", template.BaseImage, overlayPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create overlay: %w: %s", err, output)
	}

	overlay := &Overlay{
		ID:         generateID("ovl"),
		TemplateID: templateID,
		VMID:       vmID,
		Path:       overlayPath,
		CreatedAt:  time.Now(),
	}

	tm.overlays[overlay.ID] = overlay
	if err := tm.saveOverlays(); err != nil {
		os.Remove(overlayPath)
		delete(tm.overlays, overlay.ID)
		return nil, err
	}

	return overlay, nil
}

// DeleteOverlay removes an overlay file
func (tm *TemplateManager) DeleteOverlay(overlayID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	overlay, ok := tm.overlays[overlayID]
	if !ok {
		return fmt.Errorf("overlay not found: %s", overlayID)
	}

	// Delete overlay file
	if err := os.Remove(overlay.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete overlay file: %w", err)
	}

	// Mark as destroyed
	now := time.Now()
	overlay.DestroyedAt = &now

	if err := tm.saveOverlays(); err != nil {
		return err
	}

	delete(tm.overlays, overlayID)
	return nil
}

// GetOverlay returns an overlay by ID
func (tm *TemplateManager) GetOverlay(overlayID string) (*Overlay, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	overlay, ok := tm.overlays[overlayID]
	if !ok {
		return nil, fmt.Errorf("overlay not found: %s", overlayID)
	}

	return overlay, nil
}

// GetOverlaySize returns the actual disk usage of an overlay
func (tm *TemplateManager) GetOverlaySize(overlayID string) (int64, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	overlay, ok := tm.overlays[overlayID]
	if !ok {
		return 0, fmt.Errorf("overlay not found: %s", overlayID)
	}

	// qemu-img info --output=json overlay.qcow2
	cmd := exec.Command("qemu-img", "info", "--output=json", overlay.Path)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get overlay info: %w", err)
	}

	var info struct {
		ActualSize int64 `json:"actual-size"`
	}
	if err := json.Unmarshal(output, &info); err != nil {
		return 0, fmt.Errorf("failed to parse overlay info: %w", err)
	}

	return info.ActualSize, nil
}

// ListTemplates returns all templates
func (tm *TemplateManager) ListTemplates() []*pipeline.Template {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	templates := make([]*pipeline.Template, 0, len(tm.templates))
	for _, t := range tm.templates {
		templates = append(templates, t)
	}
	return templates
}

// GetTemplate returns a template by ID
func (tm *TemplateManager) GetTemplate(templateID string) (*pipeline.Template, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	template, ok := tm.templates[templateID]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	return template, nil
}

// DeleteTemplate removes a template (only if no overlays exist)
func (tm *TemplateManager) DeleteTemplate(templateID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	template, ok := tm.templates[templateID]
	if !ok {
		return fmt.Errorf("template not found: %s", templateID)
	}

	// Check for active overlays
	for _, overlay := range tm.overlays {
		if overlay.TemplateID == templateID {
			return fmt.Errorf("cannot delete template: active overlays exist")
		}
	}

	// Delete template file
	if err := os.Remove(template.BaseImage); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete template file: %w", err)
	}

	delete(tm.templates, templateID)
	return tm.saveTemplates()
}

// calculateChecksum calculates SHA256 checksum of a file
func (tm *TemplateManager) calculateChecksum(path string) (string, error) {
	cmd := exec.Command("sha256sum", path)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Output format: "checksum  filename"
	parts := splitN(string(output), " ", 2)
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid checksum output")
	}

	return parts[0], nil
}

// ImportTemplate imports an existing qcow2 file as a template
func (tm *TemplateManager) ImportTemplate(name, sourcePath string, packages []string) (*pipeline.Template, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if template already exists
	for _, t := range tm.templates {
		if t.Name == name {
			return nil, fmt.Errorf("template already exists: %s", name)
		}
	}

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source file not found: %s", sourcePath)
	}

	// Get file size
	info, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Copy to templates directory
	templatePath := filepath.Join(tm.basePath, name+".qcow2")
	cmd := exec.Command("cp", sourcePath, templatePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to copy template: %w: %s", err, output)
	}

	// Calculate checksum
	checksum, err := tm.calculateChecksum(templatePath)
	if err != nil {
		os.Remove(templatePath)
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Make read-only
	if err := os.Chmod(templatePath, 0444); err != nil {
		os.Remove(templatePath)
		return nil, fmt.Errorf("failed to make template read-only: %w", err)
	}

	template := &pipeline.Template{
		ID:        generateID("tpl"),
		Name:      name,
		BaseImage: templatePath,
		Size:      info.Size(),
		Packages:  packages,
		Checksum:  checksum,
		ReadOnly:  true,
		CreatedAt: time.Now(),
	}

	tm.templates[template.ID] = template
	if err := tm.saveTemplates(); err != nil {
		delete(tm.templates, template.ID)
		os.Remove(templatePath)
		return nil, err
	}

	return template, nil
}

// Helper functions

func generateID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, randomString(8))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func splitN(s string, sep string, n int) []string {
	result := []string{}
	start := 0
	for i := 0; i < n-1; i++ {
		idx := -1
		for j := start; j < len(s); j++ {
			if len(sep) == 1 && s[j] == sep[0] {
				idx = j
				break
			} else if len(sep) > 1 && s[j:j+len(sep)] == sep {
				idx = j
				break
			}
		}
		if idx == -1 {
			break
		}
		result = append(result, s[start:idx])
		start = idx + len(sep)
	}
	result = append(result, s[start:])
	return result
}