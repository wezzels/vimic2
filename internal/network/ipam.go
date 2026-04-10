// Package network provides IPAM (IP Address Management)
package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// IPAMManager manages IP address allocation
type IPAMManager struct {
	baseCIDR    string
	gatewayIP   string
	dns         []string
	pools       map[string]*CIDRPool // CIDR -> Pool
	allocations map[string]*IPAllocation // MAC -> Allocation
	mu          sync.RWMutex
	stateFile   string
}

// CIDRPool represents a CIDR pool
type CIDRPool struct {
	CIDR       string         `json:"cidr"`
	Gateway    string         `json:"gateway"`
	Subnet     string         `json:"subnet"`
	Mask       string         `json:"mask"`
	Start      string         `json:"start"`
	End        string         `json:"end"`
	Used       map[string]bool `json:"used"` // IP -> used
	Reserved   []string       `json:"reserved"`
	Allocations map[string]string `json:"allocations"` // IP -> MAC
}

// IPAllocation represents an IP allocation
type IPAllocation struct {
	IP          string `json:"ip"`
	MAC         string `json:"mac"`
	CIDR        string `json:"cidr"`
	VMID        string `json:"vm_id"`
	NetworkID   string `json:"network_id"`
}

// IPAMConfig represents IPAM configuration
type IPAMConfig struct {
	BaseCIDR string   `json:"base_cidr"`
	DNS      []string `json:"dns"`
}

// NewIPAMManager creates a new IPAM manager
func NewIPAMManager(config *IPAMConfig) (*IPAMManager, error) {
	im := &IPAMManager{
		baseCIDR:    config.BaseCIDR,
		dns:         config.DNS,
		pools:       make(map[string]*CIDRPool),
		allocations: make(map[string]*IPAllocation),
	}

	// Parse base CIDR
	_, ipnet, err := net.ParseCIDR(config.BaseCIDR)
	if err != nil {
		return nil, fmt.Errorf("invalid base CIDR: %w", err)
	}

	// Calculate gateway (first IP)
	im.gatewayIP = incrementIP(ipnet.IP.String())

	// Load state
	if err := im.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return im, nil
}

// loadState loads IPAM state from disk
func (im *IPAMManager) loadState() error {
	stateFile := im.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
		}
		return err
	}

	var state struct {
		Pools       map[string]*CIDRPool    `json:"pools"`
		Allocations map[string]*IPAllocation `json:"allocations"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	im.pools = state.Pools
	im.allocations = state.Allocations

	if im.pools == nil {
		im.pools = make(map[string]*CIDRPool)
	}
	if im.allocations == nil {
		im.allocations = make(map[string]*IPAllocation)
	}

	return nil
}

// saveState saves IPAM state to disk
func (im *IPAMManager) saveState() error {
	im.mu.RLock()
	defer im.mu.RUnlock()

	state := struct {
		Pools       map[string]*CIDRPool    `json:"pools"`
		Allocations map[string]*IPAllocation `json:"allocations"`
	}{
		Pools:       im.pools,
		Allocations: im.allocations,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	stateFile := im.getStateFile()
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, data, 0644)
}

// getStateFile returns the state file path
func (im *IPAMManager) getStateFile() string {
	if im.stateFile != "" {
		return im.stateFile
	}
	return filepath.Join(os.Getenv("HOME"), ".vimic2", "ipam-state.json")
}

// SetStateFile sets the state file path
func (im *IPAMManager) SetStateFile(path string) {
	im.stateFile = path
}

// Allocate allocates a new CIDR from the base CIDR
func (im *IPAMManager) Allocate() (string, string, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Parse base CIDR
	_, ipnet, err := net.ParseCIDR(im.baseCIDR)
	if err != nil {
		return "", "", fmt.Errorf("invalid base CIDR: %w", err)
	}

	// Calculate next /24 subnet
	ones, _ := ipnet.Mask.Size()
	if ones > 24 {
		return "", "", fmt.Errorf("base CIDR too small: need /24 or larger")
	}

	// Find next available /24 subnet
	for i := 0; i < (1<<(24-ones)); i++ {
		// Calculate subnet IP
		subnetIP := make(net.IP, len(ipnet.IP))
		copy(subnetIP, ipnet.IP)

		// Add offset for /24 subnet
		offset := i << (32 - 24)
		for j := 3; j >= 0; j-- {
			subnetIP[3-j] |= byte((offset >> (j * 8)) & 0xff)
		}

		cidr := fmt.Sprintf("%s/24", subnetIP.String())

		// Check if already allocated
		if _, ok := im.pools[cidr]; ok {
			continue
		}

		// Create pool
		pool := &CIDRPool{
			CIDR:        cidr,
			Gateway:     incrementIP(subnetIP.String()),
			Subnet:      subnetIP.String(),
			Mask:        "255.255.255.0",
			Start:       incrementIP(subnetIP.String(), 2),
			End:         incrementIP(subnetIP.String(), 254),
			Used:        make(map[string]bool),
			Allocations: make(map[string]string),
			Reserved: []string{
				subnetIP.String(),     // Network address
				incrementIP(subnetIP.String()),   // Gateway
				incrementIP(subnetIP.String(), 255), // Broadcast
			},
		}

		// Mark reserved IPs as used
		for _, ip := range pool.Reserved {
			pool.Used[ip] = true
		}

		im.pools[cidr] = pool

		// Save state
		if err := im.saveState(); err != nil {
			delete(im.pools, cidr)
			return "", "", fmt.Errorf("failed to save state: %w", err)
		}

		return cidr, pool.Gateway, nil
	}

	return "", "", fmt.Errorf("no available CIDR pools")
}

// Release releases a CIDR back to the pool
func (im *IPAMManager) Release(cidr string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	pool, ok := im.pools[cidr]
	if !ok {
		return fmt.Errorf("CIDR not found: %s", cidr)
	}

	// Check for active allocations
	if len(pool.Allocations) > 0 {
		return fmt.Errorf("CIDR has active allocations: %d", len(pool.Allocations))
	}

	delete(im.pools, cidr)

	// Save state
	if err := im.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// Reclaim reclaims a CIDR (marks as used)
func (im *IPAMManager) Reclaim(cidr string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if _, ok := im.pools[cidr]; ok {
		// Already exists, nothing to do
		return nil
	}

	// Create pool for reclaimed CIDR
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}

	pool := &CIDRPool{
		CIDR:        cidr,
		Gateway:     incrementIP(ipnet.IP.String()),
		Subnet:      ipnet.IP.String(),
		Mask:        "255.255.255.0",
		Start:       incrementIP(ipnet.IP.String(), 2),
		End:         incrementIP(ipnet.IP.String(), 254),
		Used:        make(map[string]bool),
		Allocations: make(map[string]string),
		Reserved: []string{
			ipnet.IP.String(),
			incrementIP(ipnet.IP.String()),
			incrementIP(ipnet.IP.String(), 255),
		},
	}

	im.pools[cidr] = pool
	return nil
}

// AllocateIP allocates an IP address from a CIDR
func (im *IPAMManager) AllocateIP(cidr, mac string) (string, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	pool, ok := im.pools[cidr]
	if !ok {
		return "", fmt.Errorf("CIDR not found: %s", cidr)
	}

	// Parse start and end
	startIP := net.ParseIP(pool.Start)
	endIP := net.ParseIP(pool.End)

	// Find available IP
	for ip := startIP; !ip.Equal(endIP); ip = incrementIPNet(ip) {
		ipStr := ip.String()
		if !pool.Used[ipStr] {
			// Allocate IP
			pool.Used[ipStr] = true
			pool.Allocations[ipStr] = mac

			// Store allocation
			im.allocations[mac] = &IPAllocation{
				IP:   ipStr,
				MAC:  mac,
				CIDR: cidr,
			}

			// Save state
			if err := im.saveState(); err != nil {
				delete(pool.Used, ipStr)
				delete(pool.Allocations, ipStr)
				delete(im.allocations, mac)
				return "", fmt.Errorf("failed to save state: %w", err)
			}

			return ipStr, nil
		}
	}

	return "", fmt.Errorf("no available IP addresses in CIDR: %s", cidr)
}

// ReleaseIP releases an IP address
func (im *IPAMManager) ReleaseIP(cidr, ip string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	pool, ok := im.pools[cidr]
	if !ok {
		return fmt.Errorf("CIDR not found: %s", cidr)
	}

	mac, ok := pool.Allocations[ip]
	if !ok {
		return fmt.Errorf("IP not allocated: %s", ip)
	}

	delete(pool.Used, ip)
	delete(pool.Allocations, ip)
	delete(im.allocations, mac)

	// Save state
	if err := im.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// GetIP returns the IP for a MAC address
func (im *IPAMManager) GetIP(mac string) (string, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	allocation, ok := im.allocations[mac]
	if !ok {
		return "", fmt.Errorf("no IP allocated for MAC: %s", mac)
	}

	return allocation.IP, nil
}

// GetMAC returns the MAC for an IP address
func (im *IPAMManager) GetMAC(cidr, ip string) (string, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	pool, ok := im.pools[cidr]
	if !ok {
		return "", fmt.Errorf("CIDR not found: %s", cidr)
	}

	mac, ok := pool.Allocations[ip]
	if !ok {
		return "", fmt.Errorf("IP not allocated: %s", ip)
	}

	return mac, nil
}

// GetGateway returns the gateway for a CIDR
func (im *IPAMManager) GetGateway(cidr string) (string, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	pool, ok := im.pools[cidr]
	if !ok {
		return "", fmt.Errorf("CIDR not found: %s", cidr)
	}

	return pool.Gateway, nil
}

// GetDNS returns the DNS servers
func (im *IPAMManager) GetDNS() []string {
	return im.dns
}

// ListPools returns all CIDR pools
func (im *IPAMManager) ListPools() []*CIDRPool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	pools := make([]*CIDRPool, 0, len(im.pools))
	for _, pool := range im.pools {
		pools = append(pools, pool)
	}
	return pools
}

// Used returns the number of used CIDRs
func (im *IPAMManager) Used() int {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return len(im.pools)
}

// Available returns the number of available CIDRs
func (im *IPAMManager) Available() int {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Calculate available /24 subnets from base CIDR
	_, ipnet, _ := net.ParseCIDR(im.baseCIDR)
	ones, _ := ipnet.Mask.Size()
	maxSubnets := 1 << (24 - ones)
	return maxSubnets - len(im.pools)
}

// Helper functions

func incrementIP(ip string, offsets ...int) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip
	}

	offset := 1
	if len(offsets) > 0 {
		offset = offsets[0]
	}

	// Convert to 4-byte representation for IPv4
	parsed = parsed.To4()
	if parsed == nil {
		return ip
	}

	result := make(net.IP, 4)
	copy(result, parsed)

	// Add offset (big-endian)
	carry := offset
	for i := 3; i >= 0; i-- {
		val := int(result[i]) + carry
		result[i] = byte(val & 0xFF)
		carry = val >> 8
	}

	return result.String()
}

func incrementIPNet(ip net.IP) net.IP {
	result := make(net.IP, len(ip))
	copy(result, ip)

	for i := 3; i >= 0; i-- {
		result[i]++
		if result[i] != 0 {
			break
		}
	}

	return result
}