// Package network provides database operations for network management
package network

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// NetworkDB implements the Database interface using SQLite
type NetworkDB struct {
	db *sql.DB
}

// NewNetworkDB creates a new network database
func NewNetworkDB(dbPath string) (*NetworkDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	ndb := &NetworkDB{db: db}
	if err := ndb.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return ndb, nil
}

// initialize creates the database schema
func (ndb *NetworkDB) initialize() error {
	schema := `
	-- Networks table
	CREATE TABLE IF NOT EXISTS networks (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL DEFAULT 'bridge',
		description TEXT,
		bridge_name TEXT,
		cidr TEXT,
		gateway TEXT,
		dns TEXT,
		vlan_id INTEGER DEFAULT 0,
		vlans TEXT,
		dhcp_enabled INTEGER DEFAULT 0,
		dhcp_start TEXT,
		dhcp_end TEXT,
		nat_enabled INTEGER DEFAULT 0,
		external_ip TEXT,
		firewall_rules TEXT,
		interfaces TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Routers table
	CREATE TABLE IF NOT EXISTS routers (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		network_id TEXT,
		enabled INTEGER DEFAULT 1,
		interfaces TEXT,
		routing_table TEXT,
		nat_rules TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (network_id) REFERENCES networks(id)
	);

	-- Firewalls table
	CREATE TABLE IF NOT EXISTS firewalls (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		network_id TEXT,
		default_policy TEXT DEFAULT 'drop',
		enabled INTEGER DEFAULT 1,
		logging INTEGER DEFAULT 0,
		rules TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (network_id) REFERENCES networks(id)
	);

	-- Tunnels table
	CREATE TABLE IF NOT EXISTS tunnels (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		protocol TEXT NOT NULL,
		local_ip TEXT NOT NULL,
		remote_ip TEXT NOT NULL,
		vni INTEGER,
		source_port INTEGER,
		dest_port INTEGER,
		network_id TEXT,
		router_id TEXT,
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (network_id) REFERENCES networks(id),
		FOREIGN KEY (router_id) REFERENCES routers(id)
	);

	-- VM Interfaces table
	CREATE TABLE IF NOT EXISTS vm_interfaces (
		id TEXT PRIMARY KEY,
		vm_id TEXT NOT NULL,
		name TEXT NOT NULL,
		mac_address TEXT,
		ip_address TEXT,
		network_id TEXT,
		vlan_id INTEGER DEFAULT 0,
		trunk_vlans TEXT,
		mtu INTEGER DEFAULT 1500,
		bandwidth INTEGER DEFAULT 0,
		state TEXT DEFAULT 'up',
		port_security INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (network_id) REFERENCES networks(id)
	);

	-- Switch ports table
	CREATE TABLE IF NOT EXISTS switch_ports (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		bridge_name TEXT NOT NULL,
		type TEXT DEFAULT 'access',
		vlan_id INTEGER DEFAULT 0,
		trunk_vlans TEXT,
		interfaces TEXT,
		tunnel_id TEXT,
		enabled INTEGER DEFAULT 1,
		FOREIGN KEY (tunnel_id) REFERENCES tunnels(id)
	);

	-- Routes table (for router routing tables)
	CREATE TABLE IF NOT EXISTS routes (
		id TEXT PRIMARY KEY,
		router_id TEXT NOT NULL,
		destination TEXT NOT NULL,
		gateway TEXT,
		interface TEXT,
		metric INTEGER DEFAULT 100,
		type TEXT DEFAULT 'static',
		enabled INTEGER DEFAULT 1,
		FOREIGN KEY (router_id) REFERENCES routers(id)
	);

	-- NAT rules table
	CREATE TABLE IF NOT EXISTS nat_rules (
		id TEXT PRIMARY KEY,
		router_id TEXT NOT NULL,
		type TEXT NOT NULL,
		source_cidr TEXT,
		dest_cidr TEXT,
		external_ip TEXT,
		external_port INTEGER,
		internal_ip TEXT,
		internal_port INTEGER,
		protocol TEXT DEFAULT 'all',
		enabled INTEGER DEFAULT 1,
		FOREIGN KEY (router_id) REFERENCES routers(id)
	);

	-- Firewall rules table
	CREATE TABLE IF NOT EXISTS firewall_rules (
		id TEXT PRIMARY KEY,
		firewall_id TEXT NOT NULL,
		name TEXT,
		direction TEXT DEFAULT 'ingress',
		protocol TEXT DEFAULT 'all',
		source_cidr TEXT,
		dest_cidr TEXT,
		source_port INTEGER,
		dest_port INTEGER,
		action TEXT DEFAULT 'drop',
		priority INTEGER DEFAULT 100,
		enabled INTEGER DEFAULT 1,
		log INTEGER DEFAULT 0,
		FOREIGN KEY (firewall_id) REFERENCES firewalls(id)
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_networks_name ON networks(name);
	CREATE INDEX IF NOT EXISTS idx_routers_network ON routers(network_id);
	CREATE INDEX IF NOT EXISTS idx_tunnels_network ON tunnels(network_id);
	CREATE INDEX IF NOT EXISTS idx_tunnels_router ON tunnels(router_id);
	CREATE INDEX IF NOT EXISTS idx_interfaces_vm ON vm_interfaces(vm_id);
	CREATE INDEX IF NOT EXISTS idx_interfaces_network ON vm_interfaces(network_id);
	CREATE INDEX IF NOT EXISTS idx_routes_router ON routes(router_id);
	CREATE INDEX IF NOT EXISTS idx_nat_router ON nat_rules(router_id);
	CREATE INDEX IF NOT EXISTS idx_rules_firewall ON firewall_rules(firewall_id);
	`

	_, err := ndb.db.Exec(schema)
	return err
}

// Network operations

func (ndb *NetworkDB) SaveNetwork(ctx context.Context, network *Network) error {
	vlansJSON, _ := json.Marshal(network.VLANs)
	firewallRulesJSON, _ := json.Marshal(network.FirewallRules)
	interfacesJSON, _ := json.Marshal(network.Interfaces)
	dnsJSON, _ := json.Marshal(network.DNS)

	query := `
		INSERT OR REPLACE INTO networks 
		(id, name, type, description, bridge_name, cidr, gateway, dns, vlan_id, vlans,
		 dhcp_enabled, dhcp_start, dhcp_end, nat_enabled, external_ip, firewall_rules,
		 interfaces, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := ndb.db.ExecContext(ctx, query,
		network.ID, network.Name, network.Type, network.Description, network.BridgeName,
		network.CIDR, network.Gateway, string(dnsJSON), network.VLANID, string(vlansJSON),
		network.DHCPEnabled, network.DHCPStart, network.DHCPEnd, network.NATEnabled,
		network.ExternalIP, string(firewallRulesJSON), string(interfacesJSON),
		network.CreatedAt, network.UpdatedAt)

	return err
}

func (ndb *NetworkDB) GetNetwork(ctx context.Context, id string) (*Network, error) {
	query := `SELECT id, name, type, description, bridge_name, cidr, gateway, dns, vlan_id, vlans,
		dhcp_enabled, dhcp_start, dhcp_end, nat_enabled, external_ip, firewall_rules,
		interfaces, created_at, updated_at FROM networks WHERE id = ?`

	row := ndb.db.QueryRowContext(ctx, query, id)
	network := &Network{}
	var dnsJSON, vlansJSON, firewallRulesJSON, interfacesJSON string

	err := row.Scan(&network.ID, &network.Name, &network.Type, &network.Description,
		&network.BridgeName, &network.CIDR, &network.Gateway, &dnsJSON, &network.VLANID,
		&vlansJSON, &network.DHCPEnabled, &network.DHCPStart, &network.DHCPEnd,
		&network.NATEnabled, &network.ExternalIP, &firewallRulesJSON, &interfacesJSON,
		&network.CreatedAt, &network.UpdatedAt)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(vlansJSON), &network.VLANs)
	json.Unmarshal([]byte(firewallRulesJSON), &network.FirewallRules)
	json.Unmarshal([]byte(interfacesJSON), &network.Interfaces)
	json.Unmarshal([]byte(dnsJSON), &network.DNS)

	return network, nil
}

func (ndb *NetworkDB) ListNetworks(ctx context.Context) ([]*Network, error) {
	query := `SELECT id, name, type, description, bridge_name, cidr, gateway, dns, vlan_id, vlans,
		dhcp_enabled, dhcp_start, dhcp_end, nat_enabled, external_ip, firewall_rules,
		interfaces, created_at, updated_at FROM networks ORDER BY name`

	rows, err := ndb.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var networks []*Network
	for rows.Next() {
		network := &Network{}
		var dnsJSON, vlansJSON, firewallRulesJSON, interfacesJSON string

		err := rows.Scan(&network.ID, &network.Name, &network.Type, &network.Description,
			&network.BridgeName, &network.CIDR, &network.Gateway, &dnsJSON, &network.VLANID,
			&vlansJSON, &network.DHCPEnabled, &network.DHCPStart, &network.DHCPEnd,
			&network.NATEnabled, &network.ExternalIP, &firewallRulesJSON, &interfacesJSON,
			&network.CreatedAt, &network.UpdatedAt)

		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(vlansJSON), &network.VLANs)
		json.Unmarshal([]byte(firewallRulesJSON), &network.FirewallRules)
		json.Unmarshal([]byte(interfacesJSON), &network.Interfaces)
		json.Unmarshal([]byte(dnsJSON), &network.DNS)

		networks = append(networks, network)
	}

	return networks, nil
}

func (ndb *NetworkDB) DeleteNetwork(ctx context.Context, id string) error {
	_, err := ndb.db.ExecContext(ctx, "DELETE FROM networks WHERE id = ?", id)
	return err
}

// Router operations

func (ndb *NetworkDB) SaveRouter(ctx context.Context, router *Router) error {
	interfacesJSON, _ := json.Marshal(router.Interfaces)
	routesJSON, _ := json.Marshal(router.RoutingTable)
	natRulesJSON, _ := json.Marshal(router.NATRules)

	query := `
		INSERT OR REPLACE INTO routers 
		(id, name, network_id, enabled, interfaces, routing_table, nat_rules, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := ndb.db.ExecContext(ctx, query,
		router.ID, router.Name, router.NetworkID, router.Enabled, string(interfacesJSON),
		string(routesJSON), string(natRulesJSON), router.CreatedAt, router.UpdatedAt)

	return err
}

func (ndb *NetworkDB) GetRouter(ctx context.Context, id string) (*Router, error) {
	query := `SELECT id, name, network_id, enabled, interfaces, routing_table, nat_rules,
		created_at, updated_at FROM routers WHERE id = ?`

	row := ndb.db.QueryRowContext(ctx, query, id)
	router := &Router{}
	var interfacesJSON, routesJSON, natRulesJSON string

	err := row.Scan(&router.ID, &router.Name, &router.NetworkID, &router.Enabled,
		&interfacesJSON, &routesJSON, &natRulesJSON, &router.CreatedAt, &router.UpdatedAt)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(interfacesJSON), &router.Interfaces)
	json.Unmarshal([]byte(routesJSON), &router.RoutingTable)
	json.Unmarshal([]byte(natRulesJSON), &router.NATRules)

	return router, nil
}

func (ndb *NetworkDB) ListRouters(ctx context.Context) ([]*Router, error) {
	query := `SELECT id, name, network_id, enabled, interfaces, routing_table, nat_rules,
		created_at, updated_at FROM routers ORDER BY name`

	rows, err := ndb.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routers []*Router
	for rows.Next() {
		router := &Router{}
		var interfacesJSON, routesJSON, natRulesJSON string

		err := rows.Scan(&router.ID, &router.Name, &router.NetworkID, &router.Enabled,
			&interfacesJSON, &routesJSON, &natRulesJSON, &router.CreatedAt, &router.UpdatedAt)

		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(interfacesJSON), &router.Interfaces)
		json.Unmarshal([]byte(routesJSON), &router.RoutingTable)
		json.Unmarshal([]byte(natRulesJSON), &router.NATRules)

		routers = append(routers, router)
	}

	return routers, nil
}

func (ndb *NetworkDB) DeleteRouter(ctx context.Context, id string) error {
	_, err := ndb.db.ExecContext(ctx, "DELETE FROM routers WHERE id = ?", id)
	return err
}

// Firewall operations

func (ndb *NetworkDB) SaveFirewall(ctx context.Context, firewall *Firewall) error {
	rulesJSON, _ := json.Marshal(firewall.Rules)

	query := `
		INSERT OR REPLACE INTO firewalls 
		(id, name, network_id, default_policy, enabled, logging, rules, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := ndb.db.ExecContext(ctx, query,
		firewall.ID, firewall.Name, firewall.NetworkID, firewall.DefaultPolicy,
		firewall.Enabled, firewall.Logging, string(rulesJSON),
		firewall.CreatedAt, firewall.UpdatedAt)

	return err
}

func (ndb *NetworkDB) GetFirewall(ctx context.Context, id string) (*Firewall, error) {
	query := `SELECT id, name, network_id, default_policy, enabled, logging, rules,
		created_at, updated_at FROM firewalls WHERE id = ?`

	row := ndb.db.QueryRowContext(ctx, query, id)
	firewall := &Firewall{}
	var rulesJSON string

	err := row.Scan(&firewall.ID, &firewall.Name, &firewall.NetworkID, &firewall.DefaultPolicy,
		&firewall.Enabled, &firewall.Logging, &rulesJSON, &firewall.CreatedAt, &firewall.UpdatedAt)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(rulesJSON), &firewall.Rules)

	return firewall, nil
}

func (ndb *NetworkDB) ListFirewalls(ctx context.Context) ([]*Firewall, error) {
	query := `SELECT id, name, network_id, default_policy, enabled, logging, rules,
		created_at, updated_at FROM firewalls ORDER BY name`

	rows, err := ndb.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var firewalls []*Firewall
	for rows.Next() {
		firewall := &Firewall{}
		var rulesJSON string

		err := rows.Scan(&firewall.ID, &firewall.Name, &firewall.NetworkID, &firewall.DefaultPolicy,
			&firewall.Enabled, &firewall.Logging, &rulesJSON, &firewall.CreatedAt, &firewall.UpdatedAt)

		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(rulesJSON), &firewall.Rules)

		firewalls = append(firewalls, firewall)
	}

	return firewalls, nil
}

func (ndb *NetworkDB) DeleteFirewall(ctx context.Context, id string) error {
	_, err := ndb.db.ExecContext(ctx, "DELETE FROM firewalls WHERE id = ?", id)
	return err
}

// Tunnel operations

func (ndb *NetworkDB) SaveTunnel(ctx context.Context, tunnel *Tunnel) error {
	query := `
		INSERT OR REPLACE INTO tunnels 
		(id, name, protocol, local_ip, remote_ip, vni, source_port, dest_port,
		 network_id, router_id, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := ndb.db.ExecContext(ctx, query,
		tunnel.ID, tunnel.Name, tunnel.Protocol, tunnel.LocalIP, tunnel.RemoteIP,
		tunnel.VNI, tunnel.SourcePort, tunnel.DestPort, tunnel.NetworkID, tunnel.RouterID,
		tunnel.Enabled, tunnel.CreatedAt, tunnel.UpdatedAt)

	return err
}

func (ndb *NetworkDB) GetTunnel(ctx context.Context, id string) (*Tunnel, error) {
	query := `SELECT id, name, protocol, local_ip, remote_ip, vni, source_port, dest_port,
		network_id, router_id, enabled, created_at, updated_at FROM tunnels WHERE id = ?`

	row := ndb.db.QueryRowContext(ctx, query, id)
	tunnel := &Tunnel{}

	err := row.Scan(&tunnel.ID, &tunnel.Name, &tunnel.Protocol, &tunnel.LocalIP,
		&tunnel.RemoteIP, &tunnel.VNI, &tunnel.SourcePort, &tunnel.DestPort,
		&tunnel.NetworkID, &tunnel.RouterID, &tunnel.Enabled, &tunnel.CreatedAt, &tunnel.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return tunnel, nil
}

func (ndb *NetworkDB) ListTunnels(ctx context.Context) ([]*Tunnel, error) {
	query := `SELECT id, name, protocol, local_ip, remote_ip, vni, source_port, dest_port,
		network_id, router_id, enabled, created_at, updated_at FROM tunnels ORDER BY name`

	rows, err := ndb.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tunnels []*Tunnel
	for rows.Next() {
		tunnel := &Tunnel{}

		err := rows.Scan(&tunnel.ID, &tunnel.Name, &tunnel.Protocol, &tunnel.LocalIP,
			&tunnel.RemoteIP, &tunnel.VNI, &tunnel.SourcePort, &tunnel.DestPort,
			&tunnel.NetworkID, &tunnel.RouterID, &tunnel.Enabled, &tunnel.CreatedAt, &tunnel.UpdatedAt)

		if err != nil {
			return nil, err
		}

		tunnels = append(tunnels, tunnel)
	}

	return tunnels, nil
}

func (ndb *NetworkDB) DeleteTunnel(ctx context.Context, id string) error {
	_, err := ndb.db.ExecContext(ctx, "DELETE FROM tunnels WHERE id = ?", id)
	return err
}

// Interface operations

func (ndb *NetworkDB) SaveInterface(ctx context.Context, iface *VMInterface) error {
	trunkVlansJSON, _ := json.Marshal(iface.TrunkVLANs)

	query := `
		INSERT OR REPLACE INTO vm_interfaces 
		(id, vm_id, name, mac_address, ip_address, network_id, vlan_id, trunk_vlans,
		 mtu, bandwidth, state, port_security, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := ndb.db.ExecContext(ctx, query,
		iface.ID, iface.VMID, iface.Name, iface.MACAddress, iface.IPAddress, iface.NetworkID,
		iface.VLANID, string(trunkVlansJSON), iface.MTU, iface.Bandwidth, iface.State,
		iface.PortSecurity, iface.CreatedAt, iface.UpdatedAt)

	return err
}

func (ndb *NetworkDB) GetInterface(ctx context.Context, id string) (*VMInterface, error) {
	query := `SELECT id, vm_id, name, mac_address, ip_address, network_id, vlan_id, trunk_vlans,
		mtu, bandwidth, state, port_security, created_at, updated_at FROM vm_interfaces WHERE id = ?`

	row := ndb.db.QueryRowContext(ctx, query, id)
	iface := &VMInterface{}
	var trunkVlansJSON string

	err := row.Scan(&iface.ID, &iface.VMID, &iface.Name, &iface.MACAddress, &iface.IPAddress,
		&iface.NetworkID, &iface.VLANID, &trunkVlansJSON, &iface.MTU, &iface.Bandwidth,
		&iface.State, &iface.PortSecurity, &iface.CreatedAt, &iface.UpdatedAt)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(trunkVlansJSON), &iface.TrunkVLANs)

	return iface, nil
}

func (ndb *NetworkDB) ListInterfaces(ctx context.Context) ([]*VMInterface, error) {
	query := `SELECT id, vm_id, name, mac_address, ip_address, network_id, vlan_id, trunk_vlans,
		mtu, bandwidth, state, port_security, created_at, updated_at FROM vm_interfaces ORDER BY vm_id, name`

	rows, err := ndb.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interfaces []*VMInterface
	for rows.Next() {
		iface := &VMInterface{}
		var trunkVlansJSON string

		err := rows.Scan(&iface.ID, &iface.VMID, &iface.Name, &iface.MACAddress, &iface.IPAddress,
			&iface.NetworkID, &iface.VLANID, &trunkVlansJSON, &iface.MTU, &iface.Bandwidth,
			&iface.State, &iface.PortSecurity, &iface.CreatedAt, &iface.UpdatedAt)

		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(trunkVlansJSON), &iface.TrunkVLANs)

		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

func (ndb *NetworkDB) DeleteInterface(ctx context.Context, id string) error {
	_, err := ndb.db.ExecContext(ctx, "DELETE FROM vm_interfaces WHERE id = ?", id)
	return err
}

// Close closes the database connection
func (ndb *NetworkDB) Close() error {
	return ndb.db.Close()
}

// Backup creates a backup of the database
func (ndb *NetworkDB) Backup(ctx context.Context, destPath string) error {
	// SQLite backup using VACUUM INTO
	_, err := ndb.db.ExecContext(ctx, fmt.Sprintf("VACUUM INTO '%s'", destPath))
	return err
}

// Restore restores the database from a backup
func (ndb *NetworkDB) Restore(ctx context.Context, srcPath string) error {
	// Close current connection
	if err := ndb.db.Close(); err != nil {
		return err
	}

	// Reopen with backup
	db, err := sql.Open("sqlite3", srcPath)
	if err != nil {
		return err
	}

	ndb.db = db
	return nil
}

// GetStats returns database statistics
func (ndb *NetworkDB) GetStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	counts := []struct {
		name  string
		table string
	}{
		{"networks", "networks"},
		{"routers", "routers"},
		{"firewalls", "firewalls"},
		{"tunnels", "tunnels"},
		{"interfaces", "vm_interfaces"},
		{"routes", "routes"},
		{"nat_rules", "nat_rules"},
		{"firewall_rules", "firewall_rules"},
	}

	for _, c := range counts {
		var count int
		err := ndb.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", c.table)).Scan(&count)
		if err != nil {
			return nil, err
		}
		stats[c.name] = count
	}

	return stats, nil
}

// Migrate runs database migrations
func (ndb *NetworkDB) Migrate(ctx context.Context) error {
	// Check current schema version
	var version int
	err := ndb.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version)
	if err != nil {
		version = 0
	}

	// Run migrations based on version
	migrations := []string{
		// Migration 1: Add support for multiple VLANs
		`ALTER TABLE networks ADD COLUMN vlans TEXT DEFAULT '[]'`,
		// Migration 2: Add MTU support
		`ALTER TABLE vm_interfaces ADD COLUMN mtu INTEGER DEFAULT 1500`,
		// Migration 3: Add bandwidth limiting
		`ALTER TABLE vm_interfaces ADD COLUMN bandwidth INTEGER DEFAULT 0`,
	}

	for i := version; i < len(migrations); i++ {
		if migrations[i] != "" {
			_, err := ndb.db.ExecContext(ctx, migrations[i])
			if err != nil {
				// Column might already exist, ignore
				if !strings.Contains(err.Error(), "duplicate column name") {
					return err
				}
			}
		}
	}

	// Update schema version
	_, err = ndb.db.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", len(migrations)))
	return err
}
