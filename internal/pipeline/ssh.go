// Package pipeline provides SSH key management
package pipeline

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/ssh"
)

// SSHKeyManager manages SSH keys for VM access
type SSHKeyManager struct {
	keyPath    string
	keyType    string
	keySize    int
	privateKey []byte
	publicKey  []byte
	mu         sync.RWMutex
}

// NewSSHKeyManager creates a new SSH key manager
func NewSSHKeyManager(keyPath, keyType string, keySize int) (*SSHKeyManager, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(keyPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create key directory: %w", err)
	}

	mgr := &SSHKeyManager{
		keyPath: keyPath,
		keyType: keyType,
		keySize: keySize,
	}

	// Load existing keys or generate new ones
	if err := mgr.loadOrGenerateKeys(); err != nil {
		return nil, fmt.Errorf("failed to load/generate keys: %w", err)
	}

	return mgr, nil
}

// loadOrGenerateKeys loads existing keys or generates new ones
func (m *SSHKeyManager) loadOrGenerateKeys() error {
	privateKeyPath := filepath.Join(m.keyPath, "id_"+m.keyType)
	publicKeyPath := privateKeyPath + ".pub"

	// Try to load existing keys
	privateData, err := ioutil.ReadFile(privateKeyPath)
	if err == nil {
		publicData, err := ioutil.ReadFile(publicKeyPath)
		if err == nil {
			m.privateKey = privateData
			m.publicKey = publicData
			return nil
		}
	}

	// Generate new keys
	return m.generateKeys()
}

// generateKeys generates a new SSH key pair
func (m *SSHKeyManager) generateKeys() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	privateKeyPath := filepath.Join(m.keyPath, "id_"+m.keyType)
	publicKeyPath := privateKeyPath + ".pub"

	var cmd *exec.Cmd

	switch m.keyType {
	case "ed25519":
		// Generate ED25519 key
		cmd = exec.Command("ssh-keygen",
			"-t", "ed25519",
			"-f", privateKeyPath,
			"-N", "", // No passphrase
			"-C", "vimic2-runner",
		)
	case "rsa":
		// Generate RSA key
		cmd = exec.Command("ssh-keygen",
			"-t", "rsa",
			"-b", fmt.Sprintf("%d", m.keySize),
			"-f", privateKeyPath,
			"-N", "", // No passphrase
			"-C", "vimic2-runner",
		)
	default:
		return fmt.Errorf("unsupported key type: %s (use ed25519 or rsa)", m.keyType)
	}

	// Run ssh-keygen
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to generate keys: %w: %s", err, output)
	}

	// Set proper permissions
	if err := os.Chmod(privateKeyPath, 0600); err != nil {
		return fmt.Errorf("failed to set private key permissions: %w", err)
	}
	if err := os.Chmod(publicKeyPath, 0644); err != nil {
		return fmt.Errorf("failed to set public key permissions: %w", err)
	}

	// Load generated keys
	privateData, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}
	publicData, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	m.privateKey = privateData
	m.publicKey = publicData

	return nil
}

// GetPrivateKey returns the private key
func (m *SSHKeyManager) GetPrivateKey() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.privateKey
}

// GetPublicKey returns the public key
func (m *SSHKeyManager) GetPublicKey() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.publicKey
}

// GetPublicKeyString returns the public key as a string
func (m *SSHKeyManager) GetPublicKeyString() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return string(m.publicKey)
}

// GetPrivateKeyPath returns the path to the private key
func (m *SSHKeyManager) GetPrivateKeyPath() string {
	return filepath.Join(m.keyPath, "id_"+m.keyType)
}

// GetPublicKeyPath returns the path to the public key
func (m *SSHKeyManager) GetPublicKeyPath() string {
	return filepath.Join(m.keyPath, "id_"+m.keyType+".pub")
}

// CopyKeyToVM copies the public key to a VM using cloud-init or virt-copy-in
func (m *SSHKeyManager) CopyKeyToVM(vmName string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create cloud-init user-data with SSH key
	userData := fmt.Sprintf(`#cloud-config
users:
  - name: root
    ssh_authorized_keys:
      - %s
`, string(m.publicKey))

	// Write user-data to temp file
	tmpFile, err := ioutil.TempFile("", "user-data-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(userData); err != nil {
		return fmt.Errorf("failed to write user-data: %w", err)
	}
	tmpFile.Close()

	// Create cloud-init ISO
	isoPath := filepath.Join("/tmp", fmt.Sprintf("%s-cloud-init.iso", vmName))
	cmd := exec.Command("cloud-localds", isoPath, tmpFile.Name())
	if _, err := cmd.CombinedOutput(); err != nil {
		// Fallback: use virt-copy-in
		return m.copyKeyWithVirtCopyIn(vmName)
	}
	defer os.Remove(isoPath)

	// Attach ISO to VM
	attachCmd := exec.Command("virsh", "attach-disk", vmName, isoPath, "hda", "--type", "cdrom", "--mode", "readonly")
	if _, err := attachCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to attach cloud-init ISO: %w", err)
	}

	return nil
}

// copyKeyWithVirtCopyIn copies the public key using virt-copy-in
func (m *SSHKeyManager) copyKeyWithVirtCopyIn(vmName string) error {
	// Write public key to temp file
	tmpFile, err := ioutil.TempFile("", "id_*.pub")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(m.publicKey); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}
	tmpFile.Close()

	// Copy key into VM
	cmd := exec.Command("virt-copy-in",
		"-a", vmName,
		tmpFile.Name(),
		"/root/.ssh/authorized_keys",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy key with virt-copy-in: %w: %s", err, output)
	}

	// Set proper permissions
	// This requires guestfish or virt-customize
	setPermCmd := exec.Command("virt-customize",
		"-a", vmName,
		"--run-command", "chmod 700 /root/.ssh && chmod 600 /root/.ssh/authorized_keys",
	)
	if output, err := setPermCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set permissions: %w: %s", err, output)
	}

	return nil
}

// GenerateCloudInitISO generates a cloud-init ISO with SSH key
func (m *SSHKeyManager) GenerateCloudInitISO(vmName, hostname string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create meta-data
	metaData := fmt.Sprintf(`instance-id: %s
local-hostname: %s
`, vmName, hostname)

	// Create user-data
	userData := fmt.Sprintf(`#cloud-config
hostname: %s
manage_etc_hosts: true
users:
  - name: root
    lock_passwd: false
    ssh_authorized_keys:
      - %s
    shell: /bin/bash

ssh_pwauth: false

runcmd:
  - systemctl enable qemu-guest-agent
  - systemctl start qemu-guest-agent
  - mkdir -p /work
`, hostname, string(m.publicKey))

	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "cloud-init-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write meta-data
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "meta-data"), []byte(metaData), 0644); err != nil {
		return "", fmt.Errorf("failed to write meta-data: %w", err)
	}

	// Write user-data
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "user-data"), []byte(userData), 0644); err != nil {
		return "", fmt.Errorf("failed to write user-data: %w", err)
	}

	// Generate ISO
	isoPath := filepath.Join("/tmp", fmt.Sprintf("%s-cloud-init.iso", vmName))
	cmd := exec.Command("genisoimage",
		"-output", isoPath,
		"-volid", "cidata",
		"-joliet", "-rock",
		filepath.Join(tmpDir, "meta-data"),
		filepath.Join(tmpDir, "user-data"),
	)
	if _, err := cmd.CombinedOutput(); err != nil {
		// Fallback to cloud-localds if genisoimage not available
		cmd = exec.Command("cloud-localds", isoPath,
			filepath.Join(tmpDir, "user-data"),
			filepath.Join(tmpDir, "meta-data"),
		)
		if _, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to create cloud-init ISO: %w", err)
		}
	}

	return isoPath, nil
}

// RotateKeys generates new keys and optionally re-deploys to VMs
func (m *SSHKeyManager) RotateKeys(vmNames []string) error {
	// Generate new keys
	if err := m.generateKeys(); err != nil {
		return fmt.Errorf("failed to generate new keys: %w", err)
	}

	// Re-deploy to specified VMs
	for _, vmName := range vmNames {
		if err := m.CopyKeyToVM(vmName); err != nil {
			return fmt.Errorf("failed to copy key to VM %s: %w", vmName, err)
		}
	}

	return nil
}

// ValidateKey validates an SSH key
func (m *SSHKeyManager) ValidateKey(privateKey []byte) error {
	// Parse private key
	_, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	return nil
}

// ParsePublicKey parses a public key and returns the key type and comment
func (m *SSHKeyManager) ParsePublicKey(publicKey []byte) (keyType, comment string, err error) {
	// Parse public key
	pubKey, comment, _, _, err := ssh.ParseAuthorizedKey(publicKey)
	if err != nil {
		return "", "", fmt.Errorf("invalid public key: %w", err)
	}

	return pubKey.Type(), comment, nil
}

// Fingerprint returns the fingerprint of the public key
func (m *SSHKeyManager) Fingerprint() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Parse public key
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(m.publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %w", err)
	}

	// Generate fingerprint (SHA256)
	fingerprint := ssh.FingerprintSHA256(pubKey)
	return fingerprint, nil
}

// ExportPrivateKeyPEM exports the private key in PEM format
func (m *SSHKeyManager) ExportPrivateKeyPEM() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// For ed25519 and rsa, the key is already in PEM format
	return m.privateKey, nil
}

// SignData signs data with the private key
func (m *SSHKeyManager) SignData(data []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Parse private key
	signer, err := ssh.ParsePrivateKey(m.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Sign data - the Sign method returns an ssh.Signature
	sig, err := signer.Sign(rand.Reader, data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// Return the signature blob
	return sig.Blob, nil
}
