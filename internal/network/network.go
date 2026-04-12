// Package network provides Open vSwitch-based network management for Vimic2
// Supports complex network topologies, VLANs, tunnels, routers, and firewalls
package network

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// NetworkType defines the type of network device
type NetworkType string

const (
	NetworkTypeBridge   NetworkType = "bridge"
	NetworkTypeRouter   NetworkType = "router"
	NetworkTypeFirewall NetworkType = "firewall"
	NetworkTypeSwitch   NetworkType = "switch"
	NetworkTypeTunnel   NetworkType = "tunnel"
	NetworkTypeVLAN     NetworkType = "vlan"
)

// TunnelProtocol defines the tunneling protocol
type TunnelProtocol string

const (
	TunnelVXLAN  TunnelProtocol = "vxlan"
	TunnelGRE    TunnelProtocol = "gre"
	TunnelGeneve TunnelProtocol = "geneve"
	TunnelGRETAP TunnelProtocol = "gretap"
	TunnelIPinIP TunnelProtocol = "ipip"
	TunnelSIT    TunnelProtocol = "sit"
)

// InterfaceState defines the state of a network interface
type InterfaceState string

const (
	InterfaceUp      InterfaceState = "up"
	InterfaceDown    InterfaceState = "down"
	InterfaceUnknown InterfaceState = "unknown"
)

// Network represents a virtual network managed by OVS
type Network struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        NetworkType `json:"type"`
	Description string      `json:"description"`

	// Bridge configuration
	BridgeName string `json:"bridge_name"`
	VLANID     int    `json:"vlan_id,omitempty"` // 0 = trunk, 1-4094 = VLAN
	VLANs      []int  `json:"vlans,omitempty"`   // Multiple VLANs for trunk

	// IP configuration
	CIDR    string   `json:"cidr"`    // Network CIDR (e.g., 10.0.0.0/24)
	Gateway string   `json:"gateway"` // Gateway IP
	DNS     []string `json:"dns"`     // DNS servers

	// DHCP configuration
	DHCPEnabled bool   `json:"dhcp_enabled"`
	DHCPStart   string `json:"dhcp_start"` // DHCP range start
	DHCPEnd     string `json:"dhcp_end"`   // DHCP range end

	// Interfaces assigned to this network
	Interfaces []string `json:"interfaces"` // Interface IDs

	// Firewall rules
	FirewallRules []FirewallRule `json:"firewall_rules,omitempty"`

	// NAT configuration
	NATEnabled bool   `json:"nat_enabled"`
	ExternalIP string `json:"external_ip,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Router represents a virtual router
type Router struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	NetworkID    string            `json:"network_id"` // Primary network
	Interfaces   []RouterInterface `json:"interfaces"`
	RoutingTable []Route           `json:"routing_table"`
	NATRules     []NATRule         `json:"nat_rules"`
	Enabled      bool              `json:"enabled"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// RouterInterface represents a router interface
type RouterInterface struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	NetworkID  string `json:"network_id"` // Connected network
	IPAddress  string `json:"ip_address"` // IP on this network
	MACAddress string `json:"mac_address"`
	VLANID     int    `json:"vlan_id,omitempty"` // VLAN tag
	Enabled    bool   `json:"enabled"`
}

// Route represents a routing table entry
type Route struct {
	ID          string `json:"id"`
	Destination string `json:"destination"` // CIDR
	Gateway     string `json:"gateway"`     // Next hop
	Interface   string `json:"interface"`   // Interface name
	Metric      int    `json:"metric"`      // Route preference
	Type        string `json:"type"`        // static, connected, ospf, bgp
	Enabled     bool   `json:"enabled"`
}

// NATRule represents a NAT rule
type NATRule struct {
	ID           string `json:"id"`
	Type         string `json:"type"`          // snat, dnat, masquerade
	SourceCIDR   string `json:"source_cidr"`   // Source network
	DestCIDR     string `json:"dest_cidr"`     // Destination network
	ExternalIP   string `json:"external_ip"`   // External IP for SNAT
	ExternalPort int    `json:"external_port"` // External port for DNAT
	InternalIP   string `json:"internal_ip"`   // Internal IP for DNAT
	InternalPort int    `json:"internal_port"` // Internal port for DNAT
	Protocol     string `json:"protocol"`      // tcp, udp, all
	Enabled      bool   `json:"enabled"`
}

// Firewall represents a virtual firewall
type Firewall struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	NetworkID     string         `json:"network_id"`
	Rules         []FirewallRule `json:"rules"`
	DefaultPolicy string         `json:"default_policy"` // accept, drop, reject
	Enabled       bool           `json:"enabled"`
	Logging       bool           `json:"logging"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Direction  string `json:"direction"`   // ingress, egress
	Protocol   string `json:"protocol"`    // tcp, udp, icmp, all
	SourceCIDR string `json:"source_cidr"` // Source network
	DestCIDR   string `json:"dest_cidr"`   // Destination network
	SourcePort int    `json:"source_port"` // Source port
	DestPort   int    `json:"dest_port"`   // Destination port
	Action     string `json:"action"`      // accept, drop, reject
	Priority   int    `json:"priority"`    // Rule priority
	Enabled    bool   `json:"enabled"`
	Log        bool   `json:"log"` // Log matches
}

// Tunnel represents a tunnel between networks/routers
type Tunnel struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Protocol   TunnelProtocol `json:"protocol"`
	LocalIP    string         `json:"local_ip"`
	RemoteIP   string         `json:"remote_ip"`
	VNI        int            `json:"vni"`         // VXLAN VNI / GRE key
	SourcePort int            `json:"source_port"` // UDP source port
	DestPort   int            `json:"dest_port"`   // UDP dest port
	NetworkID  string         `json:"network_id"`  // Associated network
	RouterID   string         `json:"router_id"`   // Associated router (optional)
	Enabled    bool           `json:"enabled"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// VMInterface represents a VM's network interface
type VMInterface struct {
	ID           string         `json:"id"`
	VMID         string         `json:"vm_id"`
	Name         string         `json:"name"` // eth0, eth1, etc.
	MACAddress   string         `json:"mac_address"`
	IPAddress    string         `json:"ip_address"`
	NetworkID    string         `json:"network_id"`  // Connected network
	VLANID       int            `json:"vlan_id"`     // VLAN tag
	TrunkVLANs   []int          `json:"trunk_vlans"` // Trunk ports
	MTU          int            `json:"mtu"`
	Bandwidth    int64          `json:"bandwidth"` // Mbps rate limit
	State        InterfaceState `json:"state"`
	PortSecurity bool           `json:"port_security"` // Enable port security
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// SwitchPort represents an OVS switch port
type SwitchPort struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	BridgeName string   `json:"bridge_name"`
	Type       string   `json:"type"` // access, trunk, tunnel
	VLANID     int      `json:"vlan_id"`
	TrunkVLANs []int    `json:"trunk_vlans"`
	Interfaces []string `json:"interfaces"` // VM interfaces
	TunnelID   string   `json:"tunnel_id"`  // For tunnel ports
	Enabled    bool     `json:"enabled"`
}

// NetworkManager manages all network operations
type NetworkManager struct {
	mu         sync.RWMutex
	networks   map[string]*Network
	routers    map[string]*Router
	firewalls  map[string]*Firewall
	tunnels    map[string]*Tunnel
	interfaces map[string]*VMInterface
	switches   map[string]*SwitchPort
	ovs        *OVSClient
	db         Database
}

// Database interface for persistence
type Database interface {
	SaveNetwork(ctx context.Context, network *Network) error
	GetNetwork(ctx context.Context, id string) (*Network, error)
	ListNetworks(ctx context.Context) ([]*Network, error)
	DeleteNetwork(ctx context.Context, id string) error

	SaveRouter(ctx context.Context, router *Router) error
	GetRouter(ctx context.Context, id string) (*Router, error)
	ListRouters(ctx context.Context) ([]*Router, error)
	DeleteRouter(ctx context.Context, id string) error

	SaveFirewall(ctx context.Context, firewall *Firewall) error
	GetFirewall(ctx context.Context, id string) (*Firewall, error)
	ListFirewalls(ctx context.Context) ([]*Firewall, error)
	DeleteFirewall(ctx context.Context, id string) error

	SaveTunnel(ctx context.Context, tunnel *Tunnel) error
	GetTunnel(ctx context.Context, id string) (*Tunnel, error)
	ListTunnels(ctx context.Context) ([]*Tunnel, error)
	DeleteTunnel(ctx context.Context, id string) error

	SaveInterface(ctx context.Context, iface *VMInterface) error
	GetInterface(ctx context.Context, id string) (*VMInterface, error)
	ListInterfaces(ctx context.Context) ([]*VMInterface, error)
	DeleteInterface(ctx context.Context, id string) error
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(db Database) *NetworkManager {
	return &NetworkManager{
		networks:   make(map[string]*Network),
		routers:    make(map[string]*Router),
		firewalls:  make(map[string]*Firewall),
		tunnels:    make(map[string]*Tunnel),
		interfaces: make(map[string]*VMInterface),
		switches:   make(map[string]*SwitchPort),
		ovs:        NewOVSClient(),
		db:         db,
	}
}

// CreateNetwork creates a new virtual network
func (nm *NetworkManager) CreateNetwork(ctx context.Context, network *Network) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Validate CIDR
	if network.CIDR != "" {
		_, _, err := net.ParseCIDR(network.CIDR)
		if err != nil {
			return fmt.Errorf("invalid CIDR: %w", err)
		}
	}

	// Set defaults
	if network.ID == "" {
		network.ID = generateID("net")
	}
	network.CreatedAt = time.Now()
	network.UpdatedAt = time.Now()

	// Create OVS bridge
	if err := nm.ovs.CreateBridge(network.BridgeName); err != nil {
		return fmt.Errorf("failed to create bridge: %w", err)
	}

	// Configure VLANs if specified
	if network.VLANID > 0 {
		if err := nm.ovs.SetBridgeVLAN(network.BridgeName, network.VLANID); err != nil {
			return fmt.Errorf("failed to set VLAN: %w", err)
		}
	}

	// Configure multiple VLANs for trunk
	if len(network.VLANs) > 0 {
		if err := nm.ovs.SetBridgeTrunk(network.BridgeName, network.VLANs); err != nil {
			return fmt.Errorf("failed to set trunk VLANs: %w", err)
		}
	}

	// Save to database
	if err := nm.db.SaveNetwork(ctx, network); err != nil {
		return fmt.Errorf("failed to save network: %w", err)
	}

	nm.networks[network.ID] = network
	return nil
}

// CreateRouter creates a new virtual router
func (nm *NetworkManager) CreateRouter(ctx context.Context, router *Router) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if router.ID == "" {
		router.ID = generateID("router")
	}
	router.CreatedAt = time.Now()
	router.UpdatedAt = time.Now()

	// Create router namespace
	if err := nm.ovs.CreateRouterNamespace(router.ID); err != nil {
		return fmt.Errorf("failed to create router namespace: %w", err)
	}

	// Create router interfaces
	for i := range router.Interfaces {
		iface := &router.Interfaces[i]
		if iface.ID == "" {
			iface.ID = generateID("iface")
		}
		if err := nm.ovs.CreateRouterInterface(router.ID, iface); err != nil {
			return fmt.Errorf("failed to create interface %s: %w", iface.Name, err)
		}
	}

	// Add initial routes
	for _, route := range router.RoutingTable {
		if err := nm.ovs.AddRoute(router.ID, route); err != nil {
			return fmt.Errorf("failed to add route: %w", err)
		}
	}

	if err := nm.db.SaveRouter(ctx, router); err != nil {
		return fmt.Errorf("failed to save router: %w", err)
	}

	nm.routers[router.ID] = router
	return nil
}

// CreateFirewall creates a new virtual firewall
func (nm *NetworkManager) CreateFirewall(ctx context.Context, firewall *Firewall) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if firewall.ID == "" {
		firewall.ID = generateID("fw")
	}
	firewall.CreatedAt = time.Now()
	firewall.UpdatedAt = time.Now()

	// Create iptables/nftables chain
	if err := nm.ovs.CreateFirewallChain(firewall.ID, firewall.DefaultPolicy); err != nil {
		return fmt.Errorf("failed to create firewall chain: %w", err)
	}

	// Add firewall rules
	for _, rule := range firewall.Rules {
		if err := nm.ovs.AddFirewallRule(firewall.ID, rule); err != nil {
			return fmt.Errorf("failed to add rule %s: %w", rule.Name, err)
		}
	}

	if err := nm.db.SaveFirewall(ctx, firewall); err != nil {
		return fmt.Errorf("failed to save firewall: %w", err)
	}

	nm.firewalls[firewall.ID] = firewall
	return nil
}

// CreateTunnel creates a new tunnel between networks/routers
func (nm *NetworkManager) CreateTunnel(ctx context.Context, tunnel *Tunnel) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if tunnel.ID == "" {
		tunnel.ID = generateID("tun")
	}
	tunnel.CreatedAt = time.Now()
	tunnel.UpdatedAt = time.Now()

	// Create OVS tunnel port
	if err := nm.ovs.CreateTunnelPort(tunnel); err != nil {
		return fmt.Errorf("failed to create tunnel: %w", err)
	}

	if err := nm.db.SaveTunnel(ctx, tunnel); err != nil {
		return fmt.Errorf("failed to save tunnel: %w", err)
	}

	nm.tunnels[tunnel.ID] = tunnel
	return nil
}

// AttachInterface attaches a VM interface to a network
func (nm *NetworkManager) AttachInterface(ctx context.Context, vmID, networkID, ifaceName string, config *VMInterface) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Get network
	network, ok := nm.networks[networkID]
	if !ok {
		return fmt.Errorf("network %s not found", networkID)
	}

	// Create OVS port
	portName := fmt.Sprintf("%s-%s", vmID, ifaceName)
	if err := nm.ovs.CreatePort(network.BridgeName, portName); err != nil {
		return fmt.Errorf("failed to create port: %w", err)
	}

	// Configure VLAN tagging
	if config.VLANID > 0 {
		if err := nm.ovs.SetPortVLAN(portName, config.VLANID); err != nil {
			return fmt.Errorf("failed to set port VLAN: %w", err)
		}
	}

	// Configure trunk VLANs
	if len(config.TrunkVLANs) > 0 {
		if err := nm.ovs.SetPortTrunk(portName, config.TrunkVLANs); err != nil {
			return fmt.Errorf("failed to set trunk VLANs: %w", err)
		}
	}

	// Set QoS rate limiting
	if config.Bandwidth > 0 {
		if err := nm.ovs.SetPortQoS(portName, config.Bandwidth); err != nil {
			return fmt.Errorf("failed to set QoS: %w", err)
		}
	}

	// Configure port security (anti-spoofing)
	if config.PortSecurity {
		if err := nm.ovs.EnablePortSecurity(portName, config.MACAddress, config.IPAddress); err != nil {
			return fmt.Errorf("failed to enable port security: %w", err)
		}
	}

	// Create interface record
	if config.ID == "" {
		config.ID = generateID("vif")
	}
	config.VMID = vmID
	config.Name = ifaceName
	config.NetworkID = networkID
	config.State = InterfaceUp
	config.UpdatedAt = time.Now()
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}

	if err := nm.db.SaveInterface(ctx, config); err != nil {
		return fmt.Errorf("failed to save interface: %w", err)
	}

	nm.interfaces[config.ID] = config
	network.Interfaces = append(network.Interfaces, config.ID)

	return nil
}

// DetachInterface detaches a VM interface from a network
func (nm *NetworkManager) DetachInterface(ctx context.Context, ifaceID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	iface, ok := nm.interfaces[ifaceID]
	if !ok {
		return fmt.Errorf("interface %s not found", ifaceID)
	}

	// Get network
	network, ok := nm.networks[iface.NetworkID]
	if !ok {
		return fmt.Errorf("network %s not found", iface.NetworkID)
	}

	// Delete OVS port
	portName := fmt.Sprintf("%s-%s", iface.VMID, iface.Name)
	if err := nm.ovs.DeletePort(network.BridgeName, portName); err != nil {
		return fmt.Errorf("failed to delete port: %w", err)
	}

	// Remove from network interfaces
	for i, id := range network.Interfaces {
		if id == ifaceID {
			network.Interfaces = append(network.Interfaces[:i], network.Interfaces[i+1:]...)
			break
		}
	}

	// Delete from database
	if err := nm.db.DeleteInterface(ctx, ifaceID); err != nil {
		return fmt.Errorf("failed to delete interface: %w", err)
	}

	delete(nm.interfaces, ifaceID)
	return nil
}

// AddRoute adds a route to a router
func (nm *NetworkManager) AddRoute(ctx context.Context, routerID string, route Route) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	router, ok := nm.routers[routerID]
	if !ok {
		return fmt.Errorf("router %s not found", routerID)
	}

	if route.ID == "" {
		route.ID = generateID("route")
	}

	if err := nm.ovs.AddRoute(routerID, route); err != nil {
		return fmt.Errorf("failed to add route: %w", err)
	}

	router.RoutingTable = append(router.RoutingTable, route)
	router.UpdatedAt = time.Now()

	if err := nm.db.SaveRouter(ctx, router); err != nil {
		return fmt.Errorf("failed to save router: %w", err)
	}

	return nil
}

// AddFirewallRule adds a rule to a firewall
func (nm *NetworkManager) AddFirewallRule(ctx context.Context, firewallID string, rule FirewallRule) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	firewall, ok := nm.firewalls[firewallID]
	if !ok {
		return fmt.Errorf("firewall %s not found", firewallID)
	}

	if rule.ID == "" {
		rule.ID = generateID("fwr")
	}

	if err := nm.ovs.AddFirewallRule(firewallID, rule); err != nil {
		return fmt.Errorf("failed to add rule: %w", err)
	}

	firewall.Rules = append(firewall.Rules, rule)
	firewall.UpdatedAt = time.Now()

	if err := nm.db.SaveFirewall(ctx, firewall); err != nil {
		return fmt.Errorf("failed to save firewall: %w", err)
	}

	return nil
}

// ListNetworks returns all networks
func (nm *NetworkManager) ListNetworks(ctx context.Context) ([]*Network, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	networks := make([]*Network, 0, len(nm.networks))
	for _, n := range nm.networks {
		networks = append(networks, n)
	}
	return networks, nil
}

// ListRouters returns all routers
func (nm *NetworkManager) ListRouters(ctx context.Context) ([]*Router, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	routers := make([]*Router, 0, len(nm.routers))
	for _, r := range nm.routers {
		routers = append(routers, r)
	}
	return routers, nil
}

// ListTunnels returns all tunnels
func (nm *NetworkManager) ListTunnels(ctx context.Context) ([]*Tunnel, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	tunnels := make([]*Tunnel, 0, len(nm.tunnels))
	for _, t := range nm.tunnels {
		tunnels = append(tunnels, t)
	}
	return tunnels, nil
}

// GetNetworkStats returns network statistics
func (nm *NetworkManager) GetNetworkStats(ctx context.Context, networkID string) (*NetworkStats, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	network, ok := nm.networks[networkID]
	if !ok {
		return nil, fmt.Errorf("network %s not found", networkID)
	}

	stats, err := nm.ovs.GetBridgeStats(network.BridgeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return stats, nil
}

// NetworkStats represents network statistics
type NetworkStats struct {
	BridgeName     string `json:"bridge_name"`
	RxBytes        int64  `json:"rx_bytes"`
	TxBytes        int64  `json:"tx_bytes"`
	RxPackets      int64  `json:"rx_packets"`
	TxPackets      int64  `json:"tx_packets"`
	RxErrors       int64  `json:"rx_errors"`
	TxErrors       int64  `json:"tx_errors"`
	ConnectedPorts int    `json:"connected_ports"`
	FlowCount      int    `json:"flow_count"`
}

// generateID generates a unique ID with prefix
func generateID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, randomString(8))
}
