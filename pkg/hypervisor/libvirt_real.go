//go:build linux && libvirt
// +build linux,libvirt

// libvirt implementation for Linux
package hypervisor

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	libvirt "github.com/libvirt/libvirt-go"
)

// LibvirtHypervisor implements Hypervisor for Linux using libvirt
type LibvirtHypervisor struct {
	conn     *libvirt.Connect
	uri      string
	storage  string // Storage pool path
	networks map[string]*libvirt.Network
}

// LibvirtNodeConfig extends NodeConfig with libvirt-specific options
type LibvirtNodeConfig struct {
	*NodeConfig
	OSVariant string `json:"os_variant"` // e.g., "ubuntu22.04"
	PoolName  string `json:"pool_name"`    // Storage pool
}

func newLibvirtHypervisor(cfg *HostConfig) (*LibvirtHypervisor, error) {
	uri := "qemu:///system"
	if cfg != nil && cfg.Address != "" {
		uri = fmt.Sprintf("qemu+%s://%s/system",
			getSSHTransport(cfg),
			net.JoinHostPort(cfg.Address, fmt.Sprintf("%d", cfg.Port)))
	}

	conn, err := libvirt.NewConnect(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}

	hv := &LibvirtHypervisor{
		conn:     conn,
		uri:      uri,
		storage:  "/var/lib/libvirt/images",
		networks: make(map[string]*libvirt.Network),
	}

	return hv, nil
}

func getSSHTransport(cfg *HostConfig) string {
	if cfg.SSHKeyPath != "" {
		return "ssh"
	}
	return "ssh"
}

const domainXMLTemplate = `<domain type='kvm'>
  <name>{{.Name}}</name>
  <memory unit='MiB'>{{.MemoryMB}}</memory>
  <currentMemory unit='MiB'>{{.MemoryMB}}</currentMemory>
  <vcpu placement='static'>{{.CPU}}</vcpu>
  <os>
    <type arch='x86_64' machine='pc-q35-rhel8.2.0'>hvm</type>
    <boot dev='hd'/>
  </os>
  <cpu mode='host-model'/>
  <clock offset='utc'>
    <timer name='rtc' tickpolicy='catchup'/>
    <timer name='pit' tickpolicy='delay'/>
    <timer name='hpet' present='no'/>
  </clock>
  <on_poweroff>destroy</on_poweroff>
  <on_reboot>restart</on_reboot>
  <on_crash>destroy</on_crash>
  <pm>
    <suspend-to-mem enabled='no'/>
    <suspend-to-disk enabled='no'/>
  </pm>
  <devices>
    <emulator>/usr/bin/qemu-system-x86_64</emulator>
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='{{.DiskPath}}'/>
      <target dev='vda' bus='virtio'/>
      <address type='drive' controller='0' bus='0' target='0' unit='0'/>
    </disk>
    <controller type='usb' index='0' model='qemu-xhci'>
      <address type='usb' bus='0' port='1'/>
    </controller>
    <controller type='sata' index='0'>
      <address type='drive' controller='0' bus='0' target='0' unit='0'/>
    </controller>
    <controller type='pci' index='0' model='pcie-root'/>
    <interface type='network'>
      <source network='{{.Network}}'/>
      <model type='virtio'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x03' function='0x0'/>
    </interface>
    <serial type='pty'>
      <target port='0'/>
    </serial>
    <console type='pty'>
      <target type='serial' port='0'/>
    </console>
    <channel type='spicevmc'>
      <target type='virtio' name='com.redhat.spice.0'/>
    </channel>
    <input type='tablet' bus='usb'>
      <address type='usb' bus='0' port='1'/>
    </input>
    <input type='mouse' bus='ps2'/>
    <input type='keyboard' bus='ps2'/>
    <graphics type='spice' autoport='yes' listen='0.0.0.0'>
      <listen type='address' address='0.0.0.0'/>
    </graphics>
    <video>
      <model type='qxl' ram='65536' vram='65536' vgamem='16384' heads='1' primary='yes'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x02' function='0x0'/>
    </video>
    <memballoon model='virtioserial'>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x08' function='0x0'/>
    </memballoon>
  </devices>
</domain>`

type domainXMLData struct {
	Name      string
	CPU       int
	MemoryMB  uint64
	DiskPath  string
	Network   string
}

func (h *LibvirtHypervisor) CreateNode(ctx context.Context, cfg *NodeConfig) (*Node, error) {
	if cfg == nil {
		return nil, fmt.Errorf("node config required")
	}

	// Generate node ID and paths
	nodeID := fmt.Sprintf("vimic2-%s", cfg.Name)
	diskPath := filepath.Join(h.storage, nodeID+".qcow2")
	netName := cfg.Network
	if netName == "" {
		netName = "default"
	}

	// Create disk image
	if err := h.createDisk(diskPath, cfg.DiskGB); err != nil {
		return nil, fmt.Errorf("failed to create disk: %w", err)
	}

	// Generate domain XML
	data := domainXMLData{
		Name:     nodeID,
		CPU:      cfg.CPU,
		MemoryMB: cfg.MemoryMB,
		DiskPath: diskPath,
		Network:  netName,
	}

	xmlBuf := &bytes.Buffer{}
	tmpl := template.Must(template.New("domain").Parse(domainXMLTemplate))
	if err := tmpl.Execute(xmlBuf, data); err != nil {
		os.Remove(diskPath)
		return nil, fmt.Errorf("failed to generate domain XML: %w", err)
	}

	// Define and start the domain
	domain, err := h.conn.DomainCreateXML(xmlBuf.String(), libvirt.DomainStartPaused)
	if err != nil {
		os.Remove(diskPath)
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}
	defer domain.Free()

	// Start the domain
	if err := domain.Resume(); err != nil {
		domain.Destroy()
		os.Remove(diskPath)
		return nil, fmt.Errorf("failed to start domain: %w", err)
	}

	// Get IP address (wait for DHCP)
	ip, err := h.waitForIP(nodeID, 60*time.Second)
	if err != nil {
		// Not fatal - domain is running
		ip = ""
	}

	// Get UUID
	uuid, _ := domain.GetUUIDString()

	return &Node{
		ID:    uuid,
		Name:  nodeID,
		State: NodeRunning,
		IP:    ip,
		Host:  h.uri,
		Config: &NodeConfig{
			Name:     cfg.Name,
			CPU:      cfg.CPU,
			MemoryMB: cfg.MemoryMB,
			DiskGB:   cfg.DiskGB,
			Image:    cfg.Image,
			Network:  netName,
		},
		Created: time.Now(),
	}, nil
}

func (h *LibvirtHypervisor) createDisk(path string, sizeGB int) error {
	// Check if disk already exists
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	// Create qcow2 image
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", "-o", fmt.Sprintf("size=%dG", sizeGB), path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("qemu-img failed: %s, %w", string(output), err)
	}

	return nil
}

func (h *LibvirtHypervisor) waitForIP(name string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		// Try to get IP via DHCP leases
		leases, err := h.conn.GetDHCPLeases(name, nil)
		if err == nil && len(leases) > 0 {
			for _, lease := range leases {
				if lease.IPAddr != "" {
					return lease.IPAddr, nil
				}
			}
		}
		
		// Also try virsh
		if ip := h.getVirshIP(name); ip != "" {
			return ip, nil
		}
		
		time.Sleep(2 * time.Second)
	}
	
	return "", fmt.Errorf("timeout waiting for IP")
}

func (h *LibvirtHypervisor) getVirshIP(name string) string {
	cmd := exec.Command("virsh", "domifaddr", name, "--source", "lease")
	var out bytes.Buffer
	cmd.Output = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	
	// Parse output - look for IP pattern
	lines := bytes.Split(out.Bytes(), []byte("\n"))
	for _, line := range lines {
		// Simple regex-like parse for IP
		fields := bytes.Fields(line)
		for i, f := range fields {
			if bytes.Contains(f, []byte("192.168.")) || bytes.Contains(f, []byte("10.")) || bytes.Contains(f, []byte("172.")) {
				// Found IP
				ip := bytes.Trim(f, "'-")
				return string(ip)
			}
			// Also check next field as it might be the IP
			if i > 0 && (bytes.Contains(fields[i-1], []byte("192.168.")) || 
			             bytes.Contains(fields[i-1], []byte("10.")) || 
			             bytes.Contains(fields[i-1], []byte("172."))) {
				return string(f)
			}
		}
	}
	return ""
}

func (h *LibvirtHypervisor) DeleteNode(ctx context.Context, id string) error {
	domain, err := h.conn.LookupDomainByUUIDString(id)
	if err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}
	defer domain.Free()

	// Get disk path before destroying
	doc, _ := domain.GetXMLDesc(0)
	var diskPath string
	parseDiskPath(doc, &diskPath)

	// Destroy and undefine
	if err := domain.Destroy(); err != nil {
		// Try undefine anyway
		domain.Undefine()
	} else {
		domain.Undefine()
	}

	// Delete disk
	if diskPath != "" {
		os.Remove(diskPath)
	}

	return nil
}

func parseDiskPath(xmlDesc string, path *string) {
	// Simple parse - look for source file=
	start := bytes.Index([]byte(xmlDesc), []byte("source file='"))
	if start == -1 {
		start = bytes.Index([]byte(xmlDesc), []byte("source file=\""))
	}
	if start == -1 {
		return
	}
	start += 13
	end := bytes.IndexByte([]byte(xmlDesc)[start:], '\'')
	if end == -1 {
		return
	}
	*path = string([]byte(xmlDesc)[start : start+end])
}

func (h *LibvirtHypervisor) StartNode(ctx context.Context, id string) error {
	domain, err := h.conn.LookupDomainByUUIDString(id)
	if err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}
	defer domain.Free()

	state, _, err := domain.GetState()
	if err != nil {
		return err
	}

	if state == libvirt.DomainRunning {
		return nil
	}

	return domain.Resume()
}

func (h *LibvirtHypervisor) StopNode(ctx context.Context, id string) error {
	domain, err := h.conn.LookupDomainByUUIDString(id)
	if err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}
	defer domain.Free()

	state, _, err := domain.GetState()
	if err != nil {
		return err
	}

	if state == libvirt.DomainShutoff {
		return nil
	}

	// Try graceful shutdown first
	domain.Shutdown()

	// Wait for shutdown with timeout
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		state, _, _ := domain.GetState()
		if state == libvirt.DomainShutoff {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Force destroy if graceful shutdown failed
	return domain.Destroy()
}

func (h *LibvirtHypervisor) RestartNode(ctx context.Context, id string) error {
	if err := h.StopNode(ctx, id); err != nil {
		return err
	}
	return h.StartNode(ctx, id)
}

func (h *LibvirtHypervisor) ListNodes(ctx context.Context) ([]*Node, error) {
	domains, err := h.conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	nodes := make([]*Node, 0, len(domains))
	for _, domain := range domains {
		node, err := h.domainToNode(&domain)
		domain.Free()
		if err == nil {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

func (h *LibvirtHypervisor) GetNode(ctx context.Context, id string) (*Node, error) {
	domain, err := h.conn.LookupDomainByUUIDString(id)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}
	defer domain.Free()

	return h.domainToNode(domain)
}

func (h *LibvirtHypervisor) domainToNode(domain *libvirt.Domain) (*Node, error) {
	name, _ := domain.GetName()
	uuid, _ := domain.GetUUIDString()
	state, _, _ := domain.GetState()

	nodeState := NodeStopped
	if state == libvirt.DomainRunning || state == libvirt.DomainPaused {
		nodeState = NodeRunning
	} else if state == libvirt.DomainShutoff {
		nodeState = NodeStopped
	} else {
		nodeState = NodeError
	}

	// Get IP if running
	ip := ""
	if state == libvirt.DomainRunning {
		ip = h.getVirshIP(name)
	}

	return &Node{
		ID:    uuid,
		Name:  name,
		State: nodeState,
		IP:    ip,
		Host:  h.uri,
		Created: time.Now(),
	}, nil
}

func (h *LibvirtHypervisor) GetNodeStatus(ctx context.Context, id string) (*NodeStatus, error) {
	domain, err := h.conn.LookupDomainByUUIDString(id)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}
	defer domain.Free()

	state, _, err := domain.GetState()
	if err != nil {
		return nil, err
	}

	info, err := domain.GetInfo()
	if err != nil {
		return nil, err
	}

	nodeState := NodeStopped
	switch state {
	case libvirt.DomainRunning:
		nodeState = NodeRunning
	case libvirt.DomainPaused:
		nodeState = NodeRunning
	case libvirt.DomainShutoff:
		nodeState = NodeStopped
	default:
		nodeState = NodeError
	}

	return &NodeStatus{
		State:       nodeState,
		Uptime:      time.Duration(info[0].CpuTime) * time.Nanosecond,
		CPUPercent:  float64(info[0].Cpu) / float64(info[0].MaxMem) * 100,
		MemUsed:     info[0].Memory * 1024,
		MemTotal:    info[0].MaxMem * 1024,
		DiskUsedGB:  0,
		DiskTotalGB: 0,
		IP:          h.getVirshIP(info[0].Name),
	}, nil
}

func (h *LibvirtHypervisor) GetMetrics(ctx context.Context, id string) (*Metrics, error) {
	status, err := h.GetNodeStatus(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		CPU:       status.CPUPercent,
		Memory:    float64(status.MemUsed) / float64(status.MemTotal) * 100,
		Disk:      0, // Would need block info
		NetworkRX: 0, // Would need stats
		NetworkTX: 0,
		Timestamp: time.Now(),
	}, nil
}

func (h *LibvirtHypervisor) Close() error {
	for _, net := range h.networks {
		net.Free()
	}
	return h.conn.Close()
}
