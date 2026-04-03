// Package pool provides VM pool management with QEMU backing files
package pool

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Template represents a VM template
type Template struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	BaseImage    string    `json:"base_image"`
	Size         int64     `json:"size"`
	OS           string    `json:"os"`
	Arch         string    `json:"arch"`
	DefaultCPU   int       `json:"default_cpu"`
	DefaultMemory int       `json:"default_memory"`
	CreatedAt    time.Time `json:"created_at"`
}

// TemplateManager manages QEMU backing file templates
type TemplateManager struct {
	basePath    string
	overlayPath string
	templates   map[string]*Template
	overlays    map[string]*Overlay
	mu          sync.RWMutex
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
		templates:   make(map[string]*Template),
		overlays:    make(map[string]*Overlay),
	}

	// Load existing templates
	if err := tm.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return tm, nil
}

func (tm *TemplateManager) loadTemplates() error {
	files, err := ioutil.ReadDir(tm.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			data, err := ioutil.ReadFile(filepath.Join(tm.basePath, file.Name()))
			if err != nil {
				continue
			}

			var template Template
			if err := json.Unmarshal(data, &template); err != nil {
				continue
			}

			tm.templates[template.ID] = &template
		}
	}

	return nil
}

func (tm *TemplateManager) saveTemplates() error {
	for id, template := range tm.templates {
		data, err := json.MarshalIndent(template, "", "  ")
		if err != nil {
			return err
		}

		path := filepath.Join(tm.basePath, id+".json")
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func (tm *TemplateManager) saveOverlays() error {
	for id, overlay := range tm.overlays {
		data, err := json.MarshalIndent(overlay, "", "  ")
		if err != nil {
			return err
		}

		path := filepath.Join(tm.overlayPath, id+".json")
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

// CreateTemplate creates a new VM template
func (tm *TemplateManager) CreateTemplate(name, baseImage string, size int64, packages []string) (*Template, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	template := &Template{
		ID:           generateID("tpl"),
		Name:         name,
		BaseImage:    baseImage,
		Size:         size,
		CreatedAt:    time.Now(),
	}

	tm.templates[template.ID] = template
	if err := tm.saveTemplates(); err != nil {
		delete(tm.templates, template.ID)
		return nil, err
	}

	return template, nil
}

// GetTemplate returns a template by ID
func (tm *TemplateManager) GetTemplate(templateID string) (*Template, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	template, ok := tm.templates[templateID]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	return template, nil
}

// ListTemplates returns all templates
func (tm *TemplateManager) ListTemplates() []*Template {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	templates := make([]*Template, 0, len(tm.templates))
	for _, template := range tm.templates {
		templates = append(templates, template)
	}

	return templates
}

// ListOverlays returns all overlays
func (tm *TemplateManager) ListOverlays() []*Overlay {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	overlays := make([]*Overlay, 0, len(tm.overlays))
	for _, overlay := range tm.overlays {
		overlays = append(overlays, overlay)
	}

	return overlays
}
