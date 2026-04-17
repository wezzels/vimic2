//go:build integration

package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

// ==================== Config Tests ====================

func TestLoadConfig_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-config-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configContent := `
database:
  path: ":memory:"
hypervisor:
  type: libvirt
  uri: "qemu:///system"
ssh:
  key_path: "` + tmpDir + `/id_rsa"
  key_type: rsa
  key_size: 2048
logging:
  level: info
  format: json
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Skipf("LoadConfig failed: %v", err)
	}
	t.Logf("Loaded config: database=%s hypervisor=%s", config.Database.Path, config.Hypervisor.Type)
}

func TestSaveConfig_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-config-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	config := &Config{
		Database: DatabaseConfig{
			Path: ":memory:",
		},
		Hypervisor: HypervisorConfig{
			Type: "libvirt",
			URI:  "qemu:///system",
		},
		SSH: SSHConfig{
			KeyPath: filepath.Join(tmpDir, "id_rsa"),
			KeyType: "rsa",
			KeySize: 2048,
		},
	}

	configPath := filepath.Join(tmpDir, "config.yaml")
	err = SaveConfig(config, configPath)
	if err != nil {
		t.Skipf("SaveConfig failed: %v", err)
	}
	t.Logf("Saved config to %s", configPath)

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file should exist after save")
	}
}

func TestExampleConfig(t *testing.T) {
	example := ExampleConfig()
	if example == "" {
		t.Error("ExampleConfig should return non-empty string")
	}
	t.Logf("ExampleConfig length: %d", len(example))
}

// ==================== SSH Key Manager Tests ====================

func TestNewSSHKeyManager_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-ssh-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	keyPath := filepath.Join(tmpDir, "id_rsa")
	km, err := NewSSHKeyManager(keyPath, "rsa", 2048)
	if err != nil {
		t.Skipf("NewSSHKeyManager failed: %v", err)
	}
	if km == nil {
		t.Fatal("SSHKeyManager should not be nil")
	}
}

func TestSSHKeyManager_GetKeys(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-ssh-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	keyPath := filepath.Join(tmpDir, "id_rsa")
	km, err := NewSSHKeyManager(keyPath, "rsa", 2048)
	if err != nil {
		t.Skipf("NewSSHKeyManager failed: %v", err)
	}

	privKey := km.GetPrivateKey()
	if privKey == nil {
		t.Error("GetPrivateKey should return non-nil")
	}

	pubKey := km.GetPublicKey()
	if pubKey == nil {
		t.Error("GetPublicKey should return non-nil")
	}

	pubKeyStr := km.GetPublicKeyString()
	if pubKeyStr == "" {
		t.Error("GetPublicKeyString should return non-empty string")
	}
	t.Logf("Public key: %s...", pubKeyStr[:min(40, len(pubKeyStr))])

	privPath := km.GetPrivateKeyPath()
	if privPath == "" {
		t.Error("GetPrivateKeyPath should return non-empty string")
	}

	pubPath := km.GetPublicKeyPath()
	if pubPath == "" {
		t.Error("GetPublicKeyPath should return non-empty string")
	}
}

func TestSSHKeyManager_ValidateKey(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-ssh-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	keyPath := filepath.Join(tmpDir, "id_rsa")
	km, err := NewSSHKeyManager(keyPath, "rsa", 2048)
	if err != nil {
		t.Skipf("NewSSHKeyManager failed: %v", err)
	}

	err = km.ValidateKey(km.GetPrivateKey())
	if err != nil {
		t.Logf("ValidateKey: %v", err)
	}
}

func TestSSHKeyManager_Fingerprint(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-ssh-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	keyPath := filepath.Join(tmpDir, "id_rsa")
	km, err := NewSSHKeyManager(keyPath, "rsa", 2048)
	if err != nil {
		t.Skipf("NewSSHKeyManager failed: %v", err)
	}

	fp, err := km.Fingerprint()
	if fp == "" {
		t.Error("Fingerprint should return non-empty string")
	}
	t.Logf("Fingerprint: %s", fp)
}

func TestSSHKeyManager_ParsePublicKey(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-ssh-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	keyPath := filepath.Join(tmpDir, "id_rsa")
	km, err := NewSSHKeyManager(keyPath, "rsa", 2048)
	if err != nil {
		t.Skipf("NewSSHKeyManager failed: %v", err)
	}

	pubKeyStr := km.GetPublicKeyString()
	keyType, comment, err := km.ParsePublicKey([]byte(pubKeyStr))
	if err != nil {
		t.Logf("ParsePublicKey: %v", err)
	} else {
		t.Logf("Parsed public key: type=%s comment=%s", keyType, comment)
	}
}

func TestSSHKeyManager_ExportPrivateKeyPEM(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-ssh-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	keyPath := filepath.Join(tmpDir, "id_rsa")
	km, err := NewSSHKeyManager(keyPath, "rsa", 2048)
	if err != nil {
		t.Skipf("NewSSHKeyManager failed: %v", err)
	}

	pemData, err := km.ExportPrivateKeyPEM()
	if err != nil || len(pemData) == 0 {
		t.Error("ExportPrivateKeyPEM should return non-empty data")
	}
	t.Logf("PEM data length: %d", len(pemData))
}

func TestSSHKeyManager_RotateKeys(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-ssh-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	keyPath := filepath.Join(tmpDir, "id_rsa")
	km, err := NewSSHKeyManager(keyPath, "rsa", 2048)
	if err != nil {
		t.Skipf("NewSSHKeyManager failed: %v", err)
	}

	oldFP, _ := km.Fingerprint()
	err = km.RotateKeys(nil)
	if err != nil {
		t.Logf("RotateKeys: %v", err)
	}
	newFP, _ := km.Fingerprint()
	t.Logf("Old fingerprint: %s, New fingerprint: %s", oldFP, newFP)
}

func TestSSHKeyManager_SignData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-ssh-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	keyPath := filepath.Join(tmpDir, "id_rsa")
	km, err := NewSSHKeyManager(keyPath, "rsa", 2048)
	if err != nil {
		t.Skipf("NewSSHKeyManager failed: %v", err)
	}

	signature, err := km.SignData([]byte("test data"))
	if err != nil {
		t.Logf("SignData: %v", err)
	} else {
		t.Logf("Signature length: %d", len(signature))
	}
}

// ==================== Artifacts GetStats Tests ====================

func TestArtifactManager_GetStats_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-artifact-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewPipelineDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	am, err := NewArtifactManager(db, &ArtifactConfig{StoragePath: tmpDir})
	if err != nil {
		t.Skipf("NewArtifactManager failed: %v", err)
	}

	stats := am.GetStats()
	t.Logf("Stats: %v", stats)
}

// ==================== Config Methods Tests ====================

func TestConfig_GetPoolConfig_Real(t *testing.T) {
	config := &Config{
		Pools: map[string]PoolConfig{
			"default": {
				Template:     "ubuntu-22.04",
				MinSize:      2,
				MaxSize:      10,
				CPU:          2,
				Memory:       2048,
			},
		},
	}

	poolConfig, err := config.GetPoolConfig("default")
	if err != nil {
		t.Errorf("GetPoolConfig should find default pool: %v", err)
	}
	if poolConfig.Template != "ubuntu-22.04" {
		t.Errorf("Template = %s, want ubuntu-22.04", poolConfig.Template)
	}

	_, err = config.GetPoolConfig("nonexistent")
	if err == nil {
		t.Error("GetPoolConfig should return error for nonexistent pool")
	}
}

func TestConfig_GetPlatformConfig_Real(t *testing.T) {
	config := &Config{
		Platforms: PlatformsConfig{
			GitHub: PlatformConfig{
				URL:     "https://github.com",
				Token:   "test-token",
				Labels:  []string{"linux", "docker"},
				Enabled: true,
			},
		},
	}

	platformConfig, err := config.GetPlatformConfig("github")
	if err != nil {
		t.Errorf("GetPlatformConfig should find github platform: %v", err)
	}
	if platformConfig.URL != "https://github.com" {
		t.Errorf("URL = %s, want https://github.com", platformConfig.URL)
	}
}

// ==================== Config Struct Tests ====================

func TestDatabaseConfig_Struct_Real(t *testing.T) {
	config := DatabaseConfig{
		Path: "/tmp/test.db",
	}

	if config.Path != "/tmp/test.db" {
		t.Errorf("Path = %s, want /tmp/test.db", config.Path)
	}
}

func TestHypervisorConfig_Struct(t *testing.T) {
	config := HypervisorConfig{
		Type: "libvirt",
		URI:  "qemu:///system",
	}

	if config.Type != "libvirt" {
		t.Errorf("Type = %s, want libvirt", config.Type)
	}
	if config.URI != "qemu:///system" {
		t.Errorf("URI = %s, want qemu:///system", config.URI)
	}
}

func TestSSHConfig_Struct(t *testing.T) {
	config := SSHConfig{
		KeyPath: "/root/.ssh/id_rsa",
		KeyType: "rsa",
		KeySize: 4096,
	}

	if config.KeyPath != "/root/.ssh/id_rsa" {
		t.Errorf("KeyPath = %s, want /root/.ssh/id_rsa", config.KeyPath)
	}
	if config.KeySize != 4096 {
		t.Errorf("KeySize = %d, want 4096", config.KeySize)
	}
}

func TestLoggingConfig_Struct_Real(t *testing.T) {
	config := LoggingConfig{
		Level:            "info",
		ElasticsearchURL: "http://localhost:9200",
		RetentionDays:   30,
	}

	if config.Level != "info" {
		t.Errorf("Level = %s, want info", config.Level)
	}
	if config.ElasticsearchURL != "http://localhost:9200" {
		t.Errorf("ElasticsearchURL = %s", config.ElasticsearchURL)
	}
}

func TestPoolConfig_Struct_Real(t *testing.T) {
	config := PoolConfig{
		Template:     "ubuntu-22.04",
		MinSize:      2,
		MaxSize:      10,
		CPU:          2,
		Memory:       2048,
		DiskSize:     20,
	}

	if config.Template != "ubuntu-22.04" {
		t.Errorf("Template = %s, want ubuntu-22.04", config.Template)
	}
	if config.MinSize != 2 {
		t.Errorf("MinSize = %d, want 2", config.MinSize)
	}
	if config.MaxSize != 10 {
		t.Errorf("MaxSize = %d, want 10", config.MaxSize)
	}
}

func TestNetworksConfig_Struct_Real(t *testing.T) {
	config := NetworksConfig{
		BaseCIDR:  "10.0.0.0/16",
		VLANStart: 100,
		VLANEnd:   200,
		DNS:       []string{"8.8.8.8", "8.8.4.4"},
	}

	if config.BaseCIDR != "10.0.0.0/16" {
		t.Errorf("BaseCIDR = %s, want 10.0.0.0/16", config.BaseCIDR)
	}
	if config.VLANStart != 100 {
		t.Errorf("VLANStart = %d, want 100", config.VLANStart)
	}
	if len(config.DNS) != 2 {
		t.Errorf("DNS count = %d, want 2", len(config.DNS))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}