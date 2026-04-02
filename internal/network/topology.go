// Package network provides topology visualization for network management
package network

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// TopologyNode represents a node in the network topology
type TopologyNode struct {
	ID     string
	Name   string
	Type   string // network, router, firewall, vm, tunnel
	X, Y   float32
	State  string // up, down, error
	Color  color.Color
}

// TopologyEdge represents a connection between nodes
type TopologyEdge struct {
	FromID string
	ToID   string
	Label  string
	Color  color.Color
}

// TopologyView provides a visual representation of network topology
type TopologyView struct {
	widget.BaseWidget

	manager   *NetworkManager
	nodes     []TopologyNode
	edges     []TopologyEdge
	scale     float32
	offsetX   float32
	offsetY   float32
	selected  string
	onSelect  func(nodeID string)

	// UI elements
	toolbar    *widget.Toolbar
	infoPanel  *widget.Label
	canvas     *fyne.Container
}

// NewTopologyView creates a new topology view
func NewTopologyView(manager *NetworkManager) *TopologyView {
	tv := &TopologyView{
		manager: manager,
		scale:   1.0,
		nodes:   make([]TopologyNode, 0),
		edges:   make([]TopologyEdge, 0),
	}
	tv.ExtendBaseWidget(tv)
	return tv
}

// CreateRenderer creates the widget renderer
func (tv *TopologyView) CreateRenderer() fyne.WidgetRenderer {
	// Create toolbar
	tv.toolbar = widget.NewToolbar(
		widget.NewToolbarAction(theme.ZoomInIcon(), func() {
			tv.scale *= 1.2
			tv.Refresh()
		}),
		widget.NewToolbarAction(theme.ZoomOutIcon(), func() {
			tv.scale /= 1.2
			tv.Refresh()
		}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			tv.RefreshTopology()
		}),
		widget.NewToolbarAction(theme.FolderNewIcon(), func() {
			// Add new node
		}),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			// Export topology
		}),
	)

	// Create info panel
	tv.infoPanel = widget.NewLabel("Select a node to view details")
	tv.infoPanel.Wrapping = fyne.TextWrapWord

	// Create canvas
	tv.canvas = container.NewWithoutLayout()

	// Main layout
	content := container.NewBorder(
		tv.toolbar,
		container.NewVScroll(tv.infoPanel),
		nil,
		nil,
		tv.canvas,
	)

	return widget.NewSimpleRenderer(content)
}

// RefreshTopology refreshes the topology from the manager
func (tv *TopologyView) RefreshTopology() {
	tv.nodes = make([]TopologyNode, 0)
	tv.edges = make([]TopologyEdge, 0)

	// Load networks as nodes (use read lock)
	tv.manager.mu.RLock()
	defer tv.manager.mu.RUnlock()

	// Add network nodes
	for _, net := range tv.manager.networks {
		node := TopologyNode{
			ID:    net.ID,
			Name:  net.Name,
			Type:  "network",
			State: "up",
			Color: tv.getNetworkColor(net),
		}
		tv.nodes = append(tv.nodes, node)
	}

	// Add router nodes
	for _, router := range tv.manager.routers {
		node := TopologyNode{
			ID:    router.ID,
			Name:  router.Name,
			Type:  "router",
			State: "up",
			Color: tv.getRouterColor(router),
		}
		tv.nodes = append(tv.nodes, node)

		// Add edges from router to networks
		for _, iface := range router.Interfaces {
			edge := TopologyEdge{
				FromID: router.ID,
				ToID:   iface.NetworkID,
				Label:  iface.Name,
				Color:  theme.PrimaryColor(),
			}
			tv.edges = append(tv.edges, edge)
		}
	}

	// Add firewall nodes
	for _, fw := range tv.manager.firewalls {
		node := TopologyNode{
			ID:    fw.ID,
			Name:  fw.Name,
			Type:  "firewall",
			State: "up",
			Color: tv.getFirewallColor(fw),
		}
		tv.nodes = append(tv.nodes, node)

		// Add edge from firewall to network
		if fw.NetworkID != "" {
			edge := TopologyEdge{
				FromID: fw.ID,
				ToID:   fw.NetworkID,
				Label:  "protects",
				Color:  color.RGBA{R: 255, G: 165, B: 0, A: 255},
			}
			tv.edges = append(tv.edges, edge)
		}
	}

	// Add tunnel nodes
	for _, tunnel := range tv.manager.tunnels {
		node := TopologyNode{
			ID:    tunnel.ID,
			Name:  tunnel.Name,
			Type:  "tunnel",
			State: "up",
			Color: tv.getTunnelColor(tunnel),
		}
		tv.nodes = append(tv.nodes, node)

		// Add edge from tunnel to network
		if tunnel.NetworkID != "" {
			edge := TopologyEdge{
				FromID: tunnel.ID,
				ToID:   tunnel.NetworkID,
				Label:  string(tunnel.Protocol),
				Color:  theme.DisabledColor(),
			}
			tv.edges = append(tv.edges, edge)
		}
	}

	// Layout nodes in a grid pattern
	tv.layoutNodes()

	// Refresh the display
	tv.Refresh()
}

// layoutNodes arranges nodes in a grid layout
func (tv *TopologyView) layoutNodes() {
	// Group nodes by type
	networkNodes := make([]TopologyNode, 0)
	routerNodes := make([]TopologyNode, 0)
	firewallNodes := make([]TopologyNode, 0)
	tunnelNodes := make([]TopologyNode, 0)

	for i := range tv.nodes {
		switch tv.nodes[i].Type {
		case "network":
			networkNodes = append(networkNodes, tv.nodes[i])
		case "router":
			routerNodes = append(routerNodes, tv.nodes[i])
		case "firewall":
			firewallNodes = append(firewallNodes, tv.nodes[i])
		case "tunnel":
			tunnelNodes = append(tunnelNodes, tv.nodes[i])
		}
	}

	// Layout networks at top
	tv.layoutGroup(networkNodes, 0, 0)

	// Layout routers in middle
	tv.layoutGroup(routerNodes, 0, 200)

	// Layout firewalls below routers
	tv.layoutGroup(firewallNodes, 0, 400)

	// Layout tunnels at bottom
	tv.layoutGroup(tunnelNodes, 0, 600)

	// Update node positions
	nodeMap := make(map[string]int)
	for i, node := range tv.nodes {
		nodeMap[node.ID] = i
	}
	for _, n := range networkNodes {
		tv.nodes[nodeMap[n.ID]].X = n.X
		tv.nodes[nodeMap[n.ID]].Y = n.Y
	}
	for _, n := range routerNodes {
		tv.nodes[nodeMap[n.ID]].X = n.X
		tv.nodes[nodeMap[n.ID]].Y = n.Y
	}
	for _, n := range firewallNodes {
		tv.nodes[nodeMap[n.ID]].X = n.X
		tv.nodes[nodeMap[n.ID]].Y = n.Y
	}
	for _, n := range tunnelNodes {
		tv.nodes[nodeMap[n.ID]].X = n.X
		tv.nodes[nodeMap[n.ID]].Y = n.Y
	}
}

// layoutGroup arranges a group of nodes in a row
func (tv *TopologyView) layoutGroup(nodes []TopologyNode, startX, startY float32) {
	spacing := float32(150)
	for i := range nodes {
		nodes[i].X = startX + float32(i)*spacing + 50
		nodes[i].Y = startY
	}
}

// getNetworkColor returns the color for a network node
func (tv *TopologyView) getNetworkColor(net *Network) color.Color {
	switch net.Type {
	case NetworkTypeBridge:
		return color.RGBA{R: 0, G: 128, B: 0, A: 255} // Green
	case NetworkTypeVLAN:
		return color.RGBA{R: 0, G: 0, B: 128, A: 255} // Dark blue
	case NetworkTypeSwitch:
		return color.RGBA{R: 128, G: 0, B: 0, A: 255} // Dark red
	default:
		return theme.PrimaryColor()
	}
}

// getRouterColor returns the color for a router node
func (tv *TopologyView) getRouterColor(router *Router) color.Color {
	if router.Enabled {
		return color.RGBA{R: 255, G: 140, B: 0, A: 255} // Orange
	}
	return theme.DisabledColor()
}

// getFirewallColor returns the color for a firewall node
func (tv *TopologyView) getFirewallColor(fw *Firewall) color.Color {
	if fw.Enabled {
		return color.RGBA{R: 220, G: 20, B: 60, A: 255} // Crimson
	}
	return theme.DisabledColor()
}

// getTunnelColor returns the color for a tunnel node
func (tv *TopologyView) getTunnelColor(tunnel *Tunnel) color.Color {
	switch tunnel.Protocol {
	case TunnelVXLAN:
		return color.RGBA{R: 138, G: 43, B: 226, A: 255} // Blue violet
	case TunnelGRE:
		return color.RGBA{R: 75, G: 0, B: 130, A: 255} // Indigo
	case TunnelGeneve:
		return color.RGBA{R: 148, G: 0, B: 211, A: 255} // Dark violet
	default:
		return theme.DisabledColor()
	}
}

// Draw draws the topology on the canvas
func (tv *TopologyView) Draw() {
	tv.canvas.Objects = nil

	// Draw edges first (below nodes)
	for _, edge := range tv.edges {
		fromNode := tv.findNode(edge.FromID)
		toNode := tv.findNode(edge.ToID)
		if fromNode == nil || toNode == nil {
			continue
		}

		// Draw line
		line := canvas.NewLine(edge.Color)
		line.Position1 = fyne.NewPos(fromNode.X*tv.scale+tv.offsetX, fromNode.Y*tv.scale+tv.offsetY)
		line.Position2 = fyne.NewPos(toNode.X*tv.scale+tv.offsetX, toNode.Y*tv.scale+tv.offsetY)
		line.StrokeWidth = 2
		tv.canvas.Add(line)

		// Draw label at midpoint
		midX := (fromNode.X + toNode.X) / 2 * tv.scale
		midY := (fromNode.Y + toNode.Y) / 2 * tv.scale
		label := canvas.NewText(edge.Label, theme.ForegroundColor())
		label.TextSize = 10
		label.Move(fyne.NewPos(midX+tv.offsetX, midY+tv.offsetY))
		tv.canvas.Add(label)
	}

	// Draw nodes
	for _, node := range tv.nodes {
		tv.drawNode(&node)
	}
}

// drawNode draws a single node on the canvas
func (tv *TopologyView) drawNode(node *TopologyNode) {
	x := node.X * tv.scale
	y := node.Y * tv.scale

	// Draw node shape based on type
	var shape fyne.CanvasObject
	nodeSize := float32(60) * tv.scale

	switch node.Type {
	case "network":
		// Rectangle for networks
		rect := canvas.NewRectangle(node.Color)
		rect.Resize(fyne.NewSize(nodeSize, nodeSize/2))
		rect.Move(fyne.NewPos(x+tv.offsetX, y+tv.offsetY))
		shape = rect

	case "router":
		// Diamond for routers (use rectangle rotated)
		rect := canvas.NewRectangle(node.Color)
		rect.Resize(fyne.NewSize(nodeSize, nodeSize))
		rect.Move(fyne.NewPos(x+tv.offsetX, y+tv.offsetY))
		shape = rect

	case "firewall":
		// Octagon for firewalls
		rect := canvas.NewRectangle(node.Color)
		rect.Resize(fyne.NewSize(nodeSize, nodeSize))
		rect.Move(fyne.NewPos(x+tv.offsetX, y+tv.offsetY))
		shape = rect

	case "tunnel":
		// Ellipse for tunnels
		circle := canvas.NewCircle(node.Color)
		circle.Resize(fyne.NewSize(nodeSize, nodeSize/2))
		circle.Move(fyne.NewPos(x+tv.offsetX, y+tv.offsetY))
		shape = circle
	}

	if shape != nil {
		tv.canvas.Add(shape)
	}

	// Draw node label
	label := canvas.NewText(node.Name, theme.ForegroundColor())
	label.TextSize = 12
	label.Alignment = fyne.TextAlignCenter
	label.Move(fyne.NewPos(x+tv.offsetX, y+nodeSize+5+tv.offsetY))
	tv.canvas.Add(label)

	// Draw status indicator
	statusColor := tv.getStatusColor(node.State)
	statusCircle := canvas.NewCircle(statusColor)
	statusCircle.Resize(fyne.NewSize(10, 10))
	statusCircle.Move(fyne.NewPos(x+nodeSize-5+tv.offsetX, y+tv.offsetY))
	tv.canvas.Add(statusCircle)
}

// getStatusColor returns the color for a status
func (tv *TopologyView) getStatusColor(status string) color.Color {
	switch status {
	case "up":
		return color.RGBA{R: 0, G: 255, B: 0, A: 255} // Green
	case "down":
		return color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red
	case "error":
		return color.RGBA{R: 255, G: 165, B: 0, A: 255} // Orange
	default:
		return theme.DisabledColor()
	}
}

// findNode finds a node by ID
func (tv *TopologyView) findNode(id string) *TopologyNode {
	for i := range tv.nodes {
		if tv.nodes[i].ID == id {
			return &tv.nodes[i]
		}
	}
	return nil
}

// SetOnSelect sets the selection callback
func (tv *TopologyView) SetOnSelect(callback func(nodeID string)) {
	tv.onSelect = callback
}

// ExportDOT exports the topology to DOT format
func (tv *TopologyView) ExportDOT() string {
	dot := "digraph network {\n"
	dot += "  rankdir=TB;\n"
	dot += "  node [shape=box];\n"

	// Add nodes
	for _, node := range tv.nodes {
		shape := "box"
		switch node.Type {
		case "network":
			shape = "box"
		case "router":
			shape = "diamond"
		case "firewall":
			shape = "octagon"
		case "tunnel":
			shape = "ellipse"
		}
		color := tv.colorToHex(node.Color)
		dot += fmt.Sprintf("  \"%s\" [label=\"%s\", shape=%s, fillcolor=\"%s\", style=filled];\n",
			node.ID, node.Name, shape, color)
	}

	// Add edges
	for _, edge := range tv.edges {
		dot += fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n",
			edge.FromID, edge.ToID, edge.Label)
	}

	dot += "}\n"
	return dot
}

// colorToHex converts color to hex string
func (tv *TopologyView) colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r/256, g/256, b/256)
}

// ExportJSON exports the topology to JSON format
func (tv *TopologyView) ExportJSON() map[string]interface{} {
	return map[string]interface{}{
		"nodes": tv.nodes,
		"edges": tv.edges,
	}
}

// ImportJSON imports topology from JSON format
func (tv *TopologyView) ImportJSON(data map[string]interface{}) error {
	// Parse nodes
	if nodes, ok := data["nodes"].([]interface{}); ok {
		tv.nodes = make([]TopologyNode, len(nodes))
		for i, n := range nodes {
			if nodeMap, ok := n.(map[string]interface{}); ok {
				tv.nodes[i] = TopologyNode{
					ID:    fmt.Sprintf("%v", nodeMap["id"]),
					Name:  fmt.Sprintf("%v", nodeMap["name"]),
					Type:  fmt.Sprintf("%v", nodeMap["type"]),
					State: fmt.Sprintf("%v", nodeMap["state"]),
				}
			}
		}
	}

	// Parse edges
	if edges, ok := data["edges"].([]interface{}); ok {
		tv.edges = make([]TopologyEdge, len(edges))
		for i, e := range edges {
			if edgeMap, ok := e.(map[string]interface{}); ok {
				tv.edges[i] = TopologyEdge{
					FromID: fmt.Sprintf("%v", edgeMap["fromId"]),
					ToID:   fmt.Sprintf("%v", edgeMap["toId"]),
					Label:  fmt.Sprintf("%v", edgeMap["label"]),
				}
			}
		}
	}

	tv.Refresh()
	return nil
}

// Tapped handles tap events on the topology
func (tv *TopologyView) Tapped(e *fyne.PointEvent) {
	// Find clicked node
	for i := range tv.nodes {
		node := &tv.nodes[i]
		nodeX := node.X * tv.scale
		nodeY := node.Y * tv.scale
		nodeSize := float32(60) * tv.scale

		// Check if click is within node bounds
		if e.Position.X >= nodeX+tv.offsetX && e.Position.X <= nodeX+nodeSize+tv.offsetX &&
			e.Position.Y >= nodeY+tv.offsetY && e.Position.Y <= nodeY+nodeSize+tv.offsetY {
			tv.selected = node.ID
			tv.showNodeInfo(node)

			if tv.onSelect != nil {
				tv.onSelect(node.ID)
			}
			break
		}
	}
}

// showNodeInfo shows information about the selected node
func (tv *TopologyView) showNodeInfo(node *TopologyNode) {
	info := fmt.Sprintf("Name: %s\nType: %s\nID: %s\nState: %s",
		node.Name, node.Type, node.ID, node.State)

	// Add type-specific info
	switch node.Type {
	case "network":
		if net, ok := tv.manager.networks[node.ID]; ok {
			info += fmt.Sprintf("\nCIDR: %s\nGateway: %s\nVLAN: %d",
				net.CIDR, net.Gateway, net.VLANID)
		}
	case "router":
		if router, ok := tv.manager.routers[node.ID]; ok {
			info += fmt.Sprintf("\nInterfaces: %d\nRoutes: %d\nNAT Rules: %d",
				len(router.Interfaces), len(router.RoutingTable), len(router.NATRules))
		}
	case "firewall":
		if fw, ok := tv.manager.firewalls[node.ID]; ok {
			info += fmt.Sprintf("\nRules: %d\nPolicy: %s\nLogging: %v",
				len(fw.Rules), fw.DefaultPolicy, fw.Logging)
		}
	case "tunnel":
		if tunnel, ok := tv.manager.tunnels[node.ID]; ok {
			info += fmt.Sprintf("\nProtocol: %s\nLocal: %s\nRemote: %s\nVNI: %d",
				tunnel.Protocol, tunnel.LocalIP, tunnel.RemoteIP, tunnel.VNI)
		}
	}

	tv.infoPanel.SetText(info)
}

// TappedSecondary handles right-click events
func (tv *TopologyView) TappedSecondary(e *fyne.PointEvent) {
	// Show context menu for node operations
}

// MouseIn handles mouse enter events
func (tv *TopologyView) MouseIn(*fyne.PointEvent) {}

// MouseOut handles mouse leave events
func (tv *TopologyView) MouseOut() {}

// LayoutNodePositions calculates optimal node positions using force-directed layout
func (tv *TopologyView) LayoutNodePositions() {
	// Simple force-directed layout algorithm
	const iterations = 100
	const k = 50.0 // Spring constant
	const repulsion = 5000.0
	const damping = 0.9

	// Initialize positions randomly
	for i := range tv.nodes {
		tv.nodes[i].X = float32(i%5) * 150 + 50
		tv.nodes[i].Y = float32(i/5) * 150 + 50
	}

	// Apply forces
	for iter := 0; iter < iterations; iter++ {
		forces := make([]struct{ fx, fy float32 }, len(tv.nodes))

		// Repulsion between all nodes
		for i := 0; i < len(tv.nodes); i++ {
			for j := i + 1; j < len(tv.nodes); j++ {
				dx := tv.nodes[i].X - tv.nodes[j].X
				dy := tv.nodes[i].Y - tv.nodes[j].Y
				dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
				if dist < 1 {
					dist = 1
				}
				force := float32(repulsion) / (dist * dist)
				forces[i].fx += force * dx / dist
				forces[i].fy += force * dy / dist
				forces[j].fx -= force * dx / dist
				forces[j].fy -= force * dy / dist
			}
		}

		// Attraction along edges
		for _, edge := range tv.edges {
			fromNode := tv.findNode(edge.FromID)
			toNode := tv.findNode(edge.ToID)
			if fromNode == nil || toNode == nil {
				continue
			}

			dx := toNode.X - fromNode.X
			dy := toNode.Y - fromNode.Y
			dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			if dist < 1 {
				dist = 1
			}
			force := k * dist / 100
			forceX := force * dx / dist
			forceY := force * dy / dist

			fromIdx := tv.findNodeIndex(edge.FromID)
			toIdx := tv.findNodeIndex(edge.ToID)
			if fromIdx >= 0 {
				forces[fromIdx].fx += forceX
				forces[fromIdx].fy += forceY
			}
			if toIdx >= 0 {
				forces[toIdx].fx -= forceX
				forces[toIdx].fy -= forceY
			}
		}

		// Apply forces with damping
		for i := range tv.nodes {
			tv.nodes[i].X += forces[i].fx * damping
			tv.nodes[i].Y += forces[i].fy * damping

			// Keep nodes in bounds
			if tv.nodes[i].X < 50 {
				tv.nodes[i].X = 50
			}
			if tv.nodes[i].X > 800 {
				tv.nodes[i].X = 800
			}
			if tv.nodes[i].Y < 50 {
				tv.nodes[i].Y = 50
			}
			if tv.nodes[i].Y > 600 {
				tv.nodes[i].Y = 600
			}
		}
	}
}

// findNodeIndex finds the index of a node by ID
func (tv *TopologyView) findNodeIndex(id string) int {
	for i := range tv.nodes {
		if tv.nodes[i].ID == id {
			return i
		}
	}
	return -1
}