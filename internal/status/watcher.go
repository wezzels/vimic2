// Package status provides real-time status updates
package status

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// Subscriber receives status updates
type Subscriber interface {
	OnNodeUpdate(*NodeUpdate)
	OnClusterUpdate(*ClusterUpdate)
}

// UpdateType indicates what changed
type UpdateType string

const (
	UpdateNode    UpdateType = "node"
	UpdateCluster UpdateType = "cluster"
	UpdateMetrics UpdateType = "metrics"
)

// NodeUpdate contains node status changes
type NodeUpdate struct {
	Type      UpdateType `json:"type"`
	NodeID    string     `json:"node_id"`
	ClusterID string     `json:"cluster_id"`
	State     string     `json:"state"`
	IP        string     `json:"ip"`
	CPU       float64    `json:"cpu"`
	Memory    float64    `json:"memory"`
	Disk      float64    `json:"disk"`
	Timestamp time.Time  `json:"timestamp"`
}

// ClusterUpdate contains cluster status changes
type ClusterUpdate struct {
	Type      UpdateType `json:"type"`
	ClusterID string     `json:"cluster_id"`
	Status    string     `json:"status"`
	NodeCount int        `json:"node_count"`
	Timestamp time.Time  `json:"timestamp"`
}

// Watcher monitors cluster and node status
type Watcher struct {
	db     *database.DB
	hosts  map[string]hypervisor.Hypervisor
	subs   []Subscriber
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewWatcher creates a new status watcher
func NewWatcher(db *database.DB, hosts map[string]hypervisor.Hypervisor) *Watcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &Watcher{
		db:     db,
		hosts:  hosts,
		subs:   make([]Subscriber, 0),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Subscribe adds a subscriber
func (w *Watcher) Subscribe(s Subscriber) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.subs = append(w.subs, s)
}

// Unsubscribe removes a subscriber
func (w *Watcher) Unsubscribe(s Subscriber) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for i, sub := range w.subs {
		if sub == s {
			w.subs = append(w.subs[:i], w.subs[i+1:]...)
			return
		}
	}
}

// Start begins watching
func (w *Watcher) Start(interval time.Duration) {
	go w.watchLoop(interval)
}

// Stop stops watching
func (w *Watcher) Stop() {
	w.cancel()
}

func (w *Watcher) watchLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkAll()
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *Watcher) checkAll() {
	clusters, err := w.db.ListClusters()
	if err != nil {
		return
	}

	for _, cluster := range clusters {
		w.checkCluster(cluster)
	}
}

func (w *Watcher) checkCluster(cluster *database.Cluster) {
	nodes, err := w.db.ListClusterNodes(cluster.ID)
	if err != nil {
		return
	}

	for _, node := range nodes {
		if node.HostID == "" {
			continue
		}

		host, ok := w.hosts[node.HostID]
		if !ok {
			continue
		}

		// Get current status
		status, err := host.GetNodeStatus(w.ctx, node.ID)
		if err != nil {
			continue
		}

		newState := "stopped"
		if status.State == hypervisor.NodeRunning {
			newState = "running"
		} else if status.State == hypervisor.NodeError {
			newState = "error"
		}

		// Check for state change
		if node.State != newState {
			w.notifyNodeUpdate(node, newState, status)
		}

		// Get metrics
		metrics, err := host.GetMetrics(w.ctx, node.ID)
		if err == nil {
			w.notifyMetrics(node, metrics)
		}
	}
}

func (w *Watcher) notifyNodeUpdate(node *database.Node, state string, status *hypervisor.NodeStatus) {
	update := &NodeUpdate{
		Type:      UpdateNode,
		NodeID:    node.ID,
		ClusterID: node.ClusterID,
		State:     state,
		Timestamp: time.Now(),
	}

	if status != nil {
		update.IP = status.IP
		update.CPU = status.CPUPercent
		update.Memory = float64(status.MemUsed) / float64(status.MemTotal) * 100
	}

	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, sub := range w.subs {
		sub.OnNodeUpdate(update)
	}

	// Update database
	if status != nil {
		w.db.UpdateNodeState(node.ID, state, status.IP)
	}
}

func (w *Watcher) notifyMetrics(node *database.Node, metrics *hypervisor.Metrics) {
	update := &NodeUpdate{
		Type:      UpdateMetrics,
		NodeID:    node.ID,
		ClusterID: node.ClusterID,
		CPU:       metrics.CPU,
		Memory:    metrics.Memory,
		Disk:      metrics.Disk,
		Timestamp: time.Now(),
	}

	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, sub := range w.subs {
		sub.OnNodeUpdate(update)
	}

	// Save to database
	w.db.SaveMetric(&database.Metric{
		NodeID:     node.ID,
		CPU:        metrics.CPU,
		Memory:     metrics.Memory,
		Disk:       metrics.Disk,
		NetworkRX:  metrics.NetworkRX,
		NetworkTX:  metrics.NetworkTX,
		RecordedAt: time.Now(),
	})
}

// WebSocketHub manages WebSocket connections for real-time updates
type WebSocketHub struct {
	clients    map[*WebSocketClient]bool
	broadcast  chan *NodeUpdate
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
}

// WebSocketClient represents a connected client
type WebSocketClient struct {
	hub        *WebSocketHub
	send       chan []byte
	nodeFilter []string // Filter updates by node IDs
}

// NewWebSocketHub creates a hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*WebSocketClient]bool),
		broadcast:  make(chan *NodeUpdate, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
	}
}

// Run starts the hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case update := <-h.broadcast:
			for client := range h.clients {
				if len(client.nodeFilter) > 0 {
					if !contains(client.nodeFilter, update.NodeID) {
						continue
					}
				}
				data, _ := json.Marshal(update)
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// Broadcast sends an update to all clients
func (h *WebSocketHub) Broadcast(update *NodeUpdate) {
	h.broadcast <- update
}

func contains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}
