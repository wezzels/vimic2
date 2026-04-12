// Package deploy provides cluster deployment orchestration
package deploy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/provisioner"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// Wizard guides users through cluster creation
type Wizard struct {
	cluster  *Cluster
	step     int
	onUpdate func(*Wizard)
	mu       sync.Mutex
}

// Cluster holds the configuration being built
type Cluster struct {
	ID         string
	Name       string
	Hosts      []*HostRef
	NodeGroups []*NodeGroup
	Network    *NetworkConfig
	Status     string
	DeployedAt time.Time
}

// HostRef references a host in the cluster
type HostRef struct {
	HostID    string
	HostName  string
	NodeCount int
}

// NodeGroup defines a group of similar nodes
type NodeGroup struct {
	Name     string
	Role     string
	Count    int
	CPU      int
	MemoryMB uint64
	DiskGB   int
	Image    string
	HostID   string // Empty = auto-select
}

// NetworkConfig holds network settings
type NetworkConfig struct {
	Type    string // nat, bridge, custom
	Name    string
	CIDR    string
	Gateway string
}

// Progress tracks deployment progress
type Progress struct {
	ClusterID     string
	TotalNodes    int
	DeployedNodes int
	CurrentNode   string
	Status        string // pending, deploying, running, error, complete
	Error         error
	StartedAt     time.Time
}

// NewWizard creates a new deployment wizard
func NewWizard() *Wizard {
	return &Wizard{
		step: 1,
		cluster: &Cluster{
			ID: uuid.New().String(),
			Network: &NetworkConfig{
				Type: "nat",
				CIDR: "192.168.100.0/24",
			},
		},
	}
}

// SetName sets the cluster name
func (w *Wizard) SetName(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cluster.Name = name
}

// SetNetwork sets network configuration
func (w *Wizard) SetNetwork(nw *NetworkConfig) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cluster.Network = nw
}

// AddNodeGroup adds a node group
func (w *Wizard) AddNodeGroup(ng *NodeGroup) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cluster.NodeGroups = append(w.cluster.NodeGroups, ng)
}

// RemoveNodeGroup removes a node group by index
func (w *Wizard) RemoveNodeGroup(idx int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if idx >= 0 && idx < len(w.cluster.NodeGroups) {
		w.cluster.NodeGroups = append(w.cluster.NodeGroups[:idx], w.cluster.NodeGroups[idx+1:]...)
	}
}

// NextStep advances to the next step
func (w *Wizard) NextStep() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.step++
	if w.onUpdate != nil {
		w.onUpdate(w)
	}
}

// PrevStep goes back one step
func (w *Wizard) PrevStep() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.step--
	if w.onUpdate != nil {
		w.onUpdate(w)
	}
}

// GetStep returns the current step
func (w *Wizard) GetStep() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.step
}

// GetCluster returns the current cluster config
func (w *Wizard) GetCluster() *Cluster {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.cluster
}

// Executor handles the actual deployment
type Executor struct {
	db          *database.DB
	hosts       map[string]hypervisor.Hypervisor
	provisioner *provisioner.Manager
}

// NewExecutor creates a new deployment executor
func NewExecutor(db *database.DB, hosts map[string]hypervisor.Hypervisor) *Executor {
	return &Executor{
		db:          db,
		hosts:       hosts,
		provisioner: provisioner.NewManager(""),
	}
}

// Execute deploys a cluster
func (e *Executor) Execute(ctx context.Context, cluster *Cluster, progress chan<- *Progress) error {
	totalNodes := 0
	for _, ng := range cluster.NodeGroups {
		totalNodes += ng.Count
	}

	baseProg := &Progress{
		ClusterID:  cluster.ID,
		TotalNodes: totalNodes,
		Status:     "deploying",
		StartedAt:  time.Now(),
	}

	// Send initial progress (copy)
	prog := *baseProg
	progress <- &prog

	// Save cluster to database
	dbCluster := &database.Cluster{
		ID:     cluster.ID,
		Name:   cluster.Name,
		Status: "deploying",
		Config: &database.ClusterConfig{
			MinNodes: totalNodes,
			MaxNodes: totalNodes,
		},
	}
	if err := e.db.SaveCluster(dbCluster); err != nil {
		return fmt.Errorf("failed to save cluster: %w", err)
	}

	deployedNodes := 0
	var lastErr error

	// Deploy each node group
	for _, ng := range cluster.NodeGroups {
		for i := 0; i < ng.Count; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			nodeName := fmt.Sprintf("%s-%s-%d", cluster.Name, ng.Role, i+1)

			// Select host
			hostID := ng.HostID
			if hostID == "" {
				hostID = e.selectBestHost()
			}

			// Get hypervisor
			hv, ok := e.hosts[hostID]
			if !ok {
				err := fmt.Errorf("host not found: %s", hostID)
				lastErr = err
				prog := *baseProg
				prog.CurrentNode = nodeName
				prog.Error = err
				prog.Status = "error"
				progress <- &prog
				continue
			}

			// Create node config
			nodeCfg := &hypervisor.NodeConfig{
				Name:     nodeName,
				CPU:      ng.CPU,
				MemoryMB: ng.MemoryMB,
				DiskGB:   ng.DiskGB,
				Image:    ng.Image,
				Network:  cluster.Network.Name,
			}

			// Create the VM
			node, err := hv.CreateNode(ctx, nodeCfg)
			if err != nil {
				err := fmt.Errorf("failed to create node %s: %w", nodeName, err)
				lastErr = err
				prog := *baseProg
				prog.CurrentNode = nodeName
				prog.Error = err
				prog.Status = "error"
				progress <- &prog
				continue
			}

			// Save node to database
			dbNode := &database.Node{
				ID:        node.ID,
				ClusterID: cluster.ID,
				HostID:    hostID,
				Name:      nodeName,
				Role:      ng.Role,
				State:     string(node.State),
				IP:        node.IP,
			}
			if err := e.db.SaveNode(dbNode); err != nil {
				// Log but don't fail
				fmt.Printf("Warning: failed to save node %s: %v\n", nodeName, err)
			}

			deployedNodes++
			prog := *baseProg
			prog.CurrentNode = nodeName
			prog.DeployedNodes = deployedNodes
			progress <- &prog
		}
	}

	// Send final progress (copy)
	finalProg := *baseProg
	if lastErr != nil {
		finalProg.Status = "error"
		finalProg.Error = lastErr
		e.db.UpdateClusterStatus(cluster.ID, "error")
	} else {
		finalProg.Status = "complete"
		finalProg.DeployedNodes = deployedNodes
		e.db.UpdateClusterStatus(cluster.ID, "running")
	}
	progress <- &finalProg

	return nil
}

// selectBestHost picks the host with most resources
func (e *Executor) selectBestHost() string {
	var bestID string
	var bestAvail float64

	for id, hv := range e.hosts {
		// Simple: just use first available
		nodes, err := hv.ListNodes(context.Background())
		if err != nil {
			continue
		}
		if len(nodes) < 10 { // Simple load balancing
			return id
		}
		if float64(len(nodes)) < bestAvail || bestID == "" {
			bestID = id
			bestAvail = float64(len(nodes))
		}
	}

	return bestID
}

// Validate checks if the cluster config is valid
func (w *Wizard) Validate() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.cluster.Name == "" {
		return fmt.Errorf("cluster name is required")
	}

	if len(w.cluster.NodeGroups) == 0 {
		return fmt.Errorf("at least one node group is required")
	}

	for _, ng := range w.cluster.NodeGroups {
		if ng.Count <= 0 {
			return fmt.Errorf("node count must be positive")
		}
		if ng.CPU <= 0 {
			ng.CPU = 2 // Default
		}
		if ng.MemoryMB == 0 {
			ng.MemoryMB = 2048 // Default 2GB
		}
		if ng.DiskGB == 0 {
			ng.DiskGB = 20 // Default 20GB
		}
		if ng.Image == "" {
			ng.Image = "ubuntu-22.04"
		}
	}

	return nil
}

// PresetTemplates returns predefined cluster templates
var PresetTemplates = map[string]*PresetTemplate{
	"dev": {
		Name:        "Development",
		Description: "Small cluster for development",
		NodeGroups: []*NodeGroup{
			{Name: "workers", Role: "worker", Count: 2, CPU: 2, MemoryMB: 2048, DiskGB: 20},
		},
	},
	"prod": {
		Name:        "Production",
		Description: "High-availability production cluster",
		NodeGroups: []*NodeGroup{
			{Name: "masters", Role: "master", Count: 3, CPU: 4, MemoryMB: 4096, DiskGB: 40},
			{Name: "workers", Role: "worker", Count: 5, CPU: 4, MemoryMB: 8192, DiskGB: 100},
		},
	},
	"db": {
		Name:        "Database",
		Description: "Database cluster with dedicated storage",
		NodeGroups: []*NodeGroup{
			{Name: "primary", Role: "database", Count: 1, CPU: 4, MemoryMB: 16384, DiskGB: 200},
			{Name: "replicas", Role: "database", Count: 2, CPU: 4, MemoryMB: 8192, DiskGB: 100},
		},
	},
}

// PresetTemplate is a predefined cluster configuration
type PresetTemplate struct {
	Name        string
	Description string
	NodeGroups  []*NodeGroup
}

// GetPreset returns a preset template by name
func GetPreset(name string) *PresetTemplate {
	if t, ok := PresetTemplates[name]; ok {
		return t
	}
	return nil
}
