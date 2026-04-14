// Package network provides UI components for network management
package network

import (
	"context"
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"image/color"
)

// NetworkTab provides the UI for network management
type NetworkTab struct {
	manager *NetworkManager
	window  fyne.Window
	app     fyne.App

	// Network list
	networkList *widget.List
	networks    []*Network
	selectedNet *Network

	// Router list
	routerList     *widget.List
	routers        []*Router
	selectedRouter *Router

	// Tunnel list
	tunnelList     *widget.List
	tunnels        []*Tunnel
	selectedTunnel *Tunnel

	// Interface list
	ifaceList     *widget.List
	interfaces    []*VMInterface
	selectedIface *VMInterface

	// Firewall list
	firewallList     *widget.List
	firewalls        []*Firewall
	selectedFirewall *Firewall

	// Topology view
	topologyCanvas *fyne.Container
	topologyNodes  map[string]*topologyNode

	// Bindings
	networkCount binding.Int
	routerCount  binding.Int
	tunnelCount  binding.Int
	ifaceCount   binding.Int
}

// NewNetworkTab creates a new network management tab
func NewNetworkTab(app fyne.App, manager *NetworkManager) *NetworkTab {
	return &NetworkTab{
		manager:      manager,
		app:          app,
		networkCount: binding.NewInt(),
		routerCount:  binding.NewInt(),
		tunnelCount:  binding.NewInt(),
		ifaceCount:   binding.NewInt(),
	}
}

// CreateUI creates the main network tab UI
func (nt *NetworkTab) CreateUI(window fyne.Window) fyne.CanvasObject {
	nt.window = window

	// Create tabs for different network views
	tabs := container.NewAppTabs(
		container.NewTabItem("Networks", nt.createNetworksTab()),
		container.NewTabItem("Routers", nt.createRoutersTab()),
		container.NewTabItem("Tunnels", nt.createTunnelsTab()),
		container.NewTabItem("Interfaces", nt.createInterfacesTab()),
		container.NewTabItem("Firewalls", nt.createFirewallsTab()),
		container.NewTabItem("Topology", nt.createTopologyTab()),
	)

	return tabs
}

// createNetworksTab creates the networks management tab
func (nt *NetworkTab) createNetworksTab() fyne.CanvasObject {
	// Network list
	nt.networkList = widget.NewList(
		func() int { return len(nt.networks) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Network Name                                        CIDR             Type             Status")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(nt.networks) {
				net := nt.networks[id]
				label := obj.(*widget.Label)
				label.SetText(fmt.Sprintf("%-40s %-16s %-16s %d interfaces",
					net.Name, net.CIDR, string(net.Type), len(net.Interfaces)))
			}
		},
	)

	nt.networkList.OnSelected = func(id widget.ListItemID) {
		if id < len(nt.networks) {
			nt.selectedNet = nt.networks[id]
		}
	}

	// Toolbar
	toolbar := container.NewHBox(
		widget.NewButton("Add Network", func() {
			nt.showCreateNetworkDialog()
		}),
		widget.NewButton("Delete", func() {
			if nt.selectedNet != nil {
				nt.deleteNetwork(nt.selectedNet.ID)
			}
		}),
		widget.NewButton("Refresh", func() {
			nt.refreshNetworks()
		}),
	)

	// Stats
	stats := container.NewHBox(
		widget.NewLabel("Networks:"),
		widget.NewLabelWithData(binding.IntToString(nt.networkCount)),
		widget.NewLabel("Routers:"),
		widget.NewLabelWithData(binding.IntToString(nt.routerCount)),
		widget.NewLabel("Tunnels:"),
		widget.NewLabelWithData(binding.IntToString(nt.tunnelCount)),
	)

	return container.NewBorder(
		container.NewVBox(stats, toolbar),
		nil,
		nil,
		nil,
		nt.networkList,
	)
}

// createRoutersTab creates the routers management tab
func (nt *NetworkTab) createRoutersTab() fyne.CanvasObject {
	// Router list
	nt.routerList = widget.NewList(
		func() int { return len(nt.routers) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Router Name              Interfaces    Routes        Status")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(nt.routers) {
				router := nt.routers[id]
				label := obj.(*widget.Label)
				status := "Stopped"
				if router.Enabled {
					status = "Running"
				}
				label.SetText(fmt.Sprintf("%-24s %-13d %-13d %s",
					router.Name, len(router.Interfaces), len(router.RoutingTable), status))
			}
		},
	)

	nt.routerList.OnSelected = func(id widget.ListItemID) {
		if id < len(nt.routers) {
			nt.selectedRouter = nt.routers[id]
			nt.showRouterDetails(nt.selectedRouter)
		}
	}

	// Toolbar
	toolbar := container.NewHBox(
		widget.NewButton("Add Router", func() {
			nt.showCreateRouterDialog()
		}),
		widget.NewButton("Delete", func() {
			if nt.selectedRouter != nil {
				nt.deleteRouter(nt.selectedRouter.ID)
			}
		}),
		widget.NewButton("Add Route", func() {
			if nt.selectedRouter != nil {
				nt.showAddRouteDialog(nt.selectedRouter.ID)
			}
		}),
		widget.NewButton("Add NAT Rule", func() {
			if nt.selectedRouter != nil {
				nt.showAddNATRuleDialog(nt.selectedRouter.ID)
			}
		}),
		widget.NewButton("Refresh", func() {
			nt.refreshRouters()
		}),
	)

	return container.NewBorder(toolbar, nil, nil, nil, nt.routerList)
}

// createTunnelsTab creates the tunnels management tab
func (nt *NetworkTab) createTunnelsTab() fyne.CanvasObject {
	// Tunnel list
	nt.tunnelList = widget.NewList(
		func() int { return len(nt.tunnels) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Tunnel Name             Protocol      Local IP          Remote IP         VNI/Key")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(nt.tunnels) {
				tunnel := nt.tunnels[id]
				label := obj.(*widget.Label)
				label.SetText(fmt.Sprintf("%-24s %-13s %-17s %-17s %d",
					tunnel.Name, string(tunnel.Protocol), tunnel.LocalIP, tunnel.RemoteIP, tunnel.VNI))
			}
		},
	)

	nt.tunnelList.OnSelected = func(id widget.ListItemID) {
		if id < len(nt.tunnels) {
			nt.selectedTunnel = nt.tunnels[id]
		}
	}

	// Toolbar
	toolbar := container.NewHBox(
		widget.NewButton("Add VXLAN Tunnel", func() {
			nt.showCreateTunnelDialog(TunnelVXLAN)
		}),
		widget.NewButton("Add GRE Tunnel", func() {
			nt.showCreateTunnelDialog(TunnelGRE)
		}),
		widget.NewButton("Add Geneve Tunnel", func() {
			nt.showCreateTunnelDialog(TunnelGeneve)
		}),
		widget.NewButton("Delete", func() {
			if nt.selectedTunnel != nil {
				nt.deleteTunnel(nt.selectedTunnel.ID)
			}
		}),
		widget.NewButton("Refresh", func() {
			nt.refreshTunnels()
		}),
	)

	return container.NewBorder(toolbar, nil, nil, nil, nt.tunnelList)
}

// createInterfacesTab creates the interfaces management tab
func (nt *NetworkTab) createInterfacesTab() fyne.CanvasObject {
	// Interface list
	nt.ifaceList = widget.NewList(
		func() int { return len(nt.interfaces) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Interface      VM ID          Network        IP Address        VLAN     Status")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(nt.interfaces) {
				iface := nt.interfaces[id]
				label := obj.(*widget.Label)
				vlanStr := "Trunk"
				if iface.VLANID > 0 {
					vlanStr = fmt.Sprintf("%d", iface.VLANID)
				}
				label.SetText(fmt.Sprintf("%-14s %-14s %-14s %-17s %-8s %s",
					iface.Name, iface.VMID, iface.NetworkID, iface.IPAddress, vlanStr, iface.State))
			}
		},
	)

	nt.ifaceList.OnSelected = func(id widget.ListItemID) {
		if id < len(nt.interfaces) {
			nt.selectedIface = nt.interfaces[id]
		}
	}

	// Toolbar
	toolbar := container.NewHBox(
		widget.NewButton("Assign to Network", func() {
			if nt.selectedIface != nil {
				nt.showAssignInterfaceDialog(nt.selectedIface.ID)
			}
		}),
		widget.NewButton("Detach", func() {
			if nt.selectedIface != nil {
				nt.detachInterface(nt.selectedIface.ID)
			}
		}),
		widget.NewButton("Add VLAN", func() {
			if nt.selectedIface != nil {
				nt.showAddVLANDialog(nt.selectedIface.ID)
			}
		}),
		widget.NewButton("Set Trunk", func() {
			if nt.selectedIface != nil {
				nt.showSetTrunkDialog(nt.selectedIface.ID)
			}
		}),
		widget.NewButton("Refresh", func() {
			nt.refreshInterfaces()
		}),
	)

	return container.NewBorder(toolbar, nil, nil, nil, nt.ifaceList)
}

// createFirewallsTab creates the firewalls management tab
func (nt *NetworkTab) createFirewallsTab() fyne.CanvasObject {
	firewallList := widget.NewList(
		func() int { return len(nt.firewalls) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Firewall Name"),
				widget.NewLabel("Rules"),
				widget.NewLabel("Policy"),
				widget.NewLabel("Status"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(nt.firewalls) {
				fw := nt.firewalls[id]
				hbox := obj.(*fyne.Container)
				labels := hbox.Objects
				labels[0].(*widget.Label).SetText(fw.Name)
				labels[1].(*widget.Label).SetText(strconv.Itoa(len(fw.Rules)))
				labels[2].(*widget.Label).SetText(fw.DefaultPolicy)
				labels[3].(*widget.Label).SetText("active")
			}
		},
	)

	nt.firewallList = firewallList

	firewallList.OnSelected = func(id widget.ListItemID) {
		if id < len(nt.firewalls) {
			nt.selectedFirewall = nt.firewalls[id]
		}
	}

	toolbar := container.NewHBox(
		widget.NewButton("Add Firewall", func() {
			nt.showCreateFirewallDialog()
		}),
		widget.NewButton("Add Rule", func() {
			if nt.selectedFirewall != nil {
				nt.showAddFirewallRuleDialog(nt.selectedFirewall.ID)
			}
		}),
		widget.NewButton("Delete", func() {
			if nt.selectedFirewall != nil {
				nt.deleteFirewall(nt.selectedFirewall.ID)
			}
		}),
		widget.NewButton("Refresh", func() {
			nt.refreshFirewalls()
		}),
	)

	return container.NewBorder(toolbar, nil, nil, nil, firewallList)
}

// createTopologyTab creates the network topology view
func (nt *NetworkTab) createTopologyTab() fyne.CanvasObject {
	// Create topology canvas with nodes
	nt.topologyCanvas = container.NewWithoutLayout()
	nt.topologyNodes = make(map[string]*topologyNode)

	// Build initial topology
	nt.buildTopologyView()

	// Add scrollable container
	scroll := container.NewScroll(nt.topologyCanvas)
	scroll.SetMinSize(fyne.NewSize(600, 400))

	toolbar := container.NewHBox(
		widget.NewButton("Add Bridge", func() {
			nt.showCreateBridgeDialog()
		}),
		widget.NewButton("Add Router", func() {
			nt.showCreateRouterDialog()
		}),
		widget.NewButton("Connect", func() {
			nt.showConnectDialog()
		}),
		widget.NewButton("Refresh", func() {
			nt.buildTopologyView()
		}),
		widget.NewButton("Export", func() {
			nt.exportTopology()
		}),
		widget.NewButton("Import", func() {
			nt.importTopology()
		}),
	)

	return container.NewBorder(toolbar, nil, nil, nil, scroll)
}

// topologyNode represents a node in the topology view
type topologyNode struct {
	id     string
	name   string
	nodeType string // "network", "router", "vm"
	x, y   float32
	object fyne.CanvasObject
}

// buildTopologyView rebuilds the topology visualization
func (nt *NetworkTab) buildTopologyView() {
	// Clear existing content
	nt.topologyCanvas.Objects = nil
	nt.topologyNodes = make(map[string]*topologyNode)

	// Layout positions
	networkX := float32(50.0)
	routerX := float32(250.0)
	networkY := float32(50.0)
	routerY := float32(150.0)

	// Add networks
	for i, net := range nt.networks {
		yPos := networkY + float32(i)*80
		node := nt.createTopologyNode(net.ID, net.Name, "network", networkX, yPos, color.RGBA{R: 100, G: 149, B: 237, A: 255})
		nt.topologyNodes[net.ID] = node
	}

	// Add routers
	for i, router := range nt.routers {
		yPos := routerY + float32(i)*80
		node := nt.createTopologyNode(router.ID, router.Name, "router", routerX, yPos, color.RGBA{R: 255, G: 165, B: 0, A: 255})
		nt.topologyNodes[router.ID] = node
	}

	// Draw connections between networks and routers
	nt.drawTopologyConnections()
}

// createTopologyNode creates a visual node in the topology
func (nt *NetworkTab) createTopologyNode(id, name, nodeType string, x, y float32, col color.RGBA) *topologyNode {
	// Create node box
	box := canvas.NewRectangle(col)
	box.StrokeColor = color.RGBA{R: 50, G: 50, B: 50, A: 255}
	box.StrokeWidth = 2
	box.Resize(fyne.NewSize(120, 50))
	box.Move(fyne.NewPos(x, y))

	// Create label
	label := widget.NewLabel(name)
	label.Alignment = fyne.TextAlignCenter
	label.Move(fyne.NewPos(x+10, y+15))
	label.Resize(fyne.NewSize(100, 20))

	// Type indicator
	typeLabel := widget.NewLabel(nodeType)
	typeLabel.Alignment = fyne.TextAlignCenter
	typeLabel.TextStyle = fyne.TextStyle{Bold: true}
	typeLabel.Move(fyne.NewPos(x+10, y+35))
	typeLabel.Resize(fyne.NewSize(100, 12))

	// Add to canvas
	nt.topologyCanvas.Add(box)
	nt.topologyCanvas.Add(label)
	nt.topologyCanvas.Add(typeLabel)

	return &topologyNode{
		id:       id,
		name:     name,
		nodeType: nodeType,
		x:        x,
		y:        y,
		object:   box,
	}
}

// drawTopologyConnections draws lines between connected nodes
func (nt *NetworkTab) drawTopologyConnections() {
	// Draw connections from networks to routers
	for _, net := range nt.networks {
		// Find which router this network connects to
		for _, router := range nt.routers {
			for _, routerIface := range router.Interfaces {
				if routerIface.NetworkID == net.ID {
					nt.drawConnection(net.ID, router.ID)
					break // Only draw one connection per network-router pair
				}
			}
		}
	}
}

// drawConnection draws a line between two nodes
func (nt *NetworkTab) drawConnection(fromID, toID string) {
	fromNode, ok1 := nt.topologyNodes[fromID]
	toNode, ok2 := nt.topologyNodes[toID]
	if !ok1 || !ok2 {
		return
	}

	// Calculate line from center of from node to center of to node
	x1 := fromNode.x + 60 // Center of 120px wide node
	y1 := fromNode.y + 25 // Center of 50px tall node
	x2 := toNode.x + 60
	y2 := toNode.y + 25

	line := canvas.NewLine(color.RGBA{R: 100, G: 100, B: 100, A: 255})
	line.StrokeColor = color.RGBA{R: 100, G: 100, B: 100, A: 255}
	line.StrokeWidth = 2
	line.Position1 = fyne.NewPos(x1, y1)
	line.Position2 = fyne.NewPos(x2, y2)

	nt.topologyCanvas.Add(line)
}

// Refresh rebuilds the topology view with current data
func (nt *NetworkTab) refreshTopology() {
	nt.refreshNetworks()
	nt.refreshRouters()
	nt.buildTopologyView()
	nt.topologyCanvas.Refresh()
}

// Dialog implementations

func (nt *NetworkTab) showCreateNetworkDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("network-1")

	typeSelect := widget.NewSelect([]string{
		string(NetworkTypeBridge),
		string(NetworkTypeVLAN),
		string(NetworkTypeSwitch),
	}, nil)
	typeSelect.SetSelected(string(NetworkTypeBridge))

	cidrEntry := widget.NewEntry()
	cidrEntry.SetPlaceHolder("10.0.0.0/24")

	gatewayEntry := widget.NewEntry()
	gatewayEntry.SetPlaceHolder("10.0.0.1")

	vlanEntry := widget.NewEntry()
	vlanEntry.SetPlaceHolder("0 (Trunk) or 1-4094")

	dhcpCheck := widget.NewCheck("Enable DHCP", nil)
	dhcpStartEntry := widget.NewEntry()
	dhcpStartEntry.SetPlaceHolder("10.0.0.100")
	dhcpEndEntry := widget.NewEntry()
	dhcpEndEntry.SetPlaceHolder("10.0.0.200")

	natCheck := widget.NewCheck("Enable NAT", nil)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name", Widget: nameEntry},
			{Text: "Type", Widget: typeSelect},
			{Text: "CIDR", Widget: cidrEntry},
			{Text: "Gateway", Widget: gatewayEntry},
			{Text: "VLAN ID", Widget: vlanEntry},
			{Text: "", Widget: dhcpCheck},
			{Text: "DHCP Start", Widget: dhcpStartEntry},
			{Text: "DHCP End", Widget: dhcpEndEntry},
			{Text: "", Widget: natCheck},
		},
		OnSubmit: func() {
			network := &Network{
				Name:        nameEntry.Text,
				Type:        NetworkType(typeSelect.Selected),
				CIDR:        cidrEntry.Text,
				Gateway:     gatewayEntry.Text,
				DHCPEnabled: dhcpCheck.Checked,
				DHCPStart:   dhcpStartEntry.Text,
				DHCPEnd:     dhcpEndEntry.Text,
				NATEnabled:  natCheck.Checked,
			}
			if vlanEntry.Text != "" {
				vlan, _ := strconv.Atoi(vlanEntry.Text)
				network.VLANID = vlan
			}
			if err := nt.manager.CreateNetwork(context.Background(), network); err != nil {
				dialog.ShowError(err, nt.window)
				return
			}
			nt.refreshNetworks()
		},
	}

	dialog.ShowForm("Create Network", "Create", "Cancel", form.Items, func(confirm bool) {
		if confirm {
			form.OnSubmit()
		}
	}, nt.window)
}

func (nt *NetworkTab) showCreateRouterDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("router-1")

	interfaceCount := widget.NewEntry()
	interfaceCount.SetPlaceHolder("2")
	interfaceCount.SetText("2")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name", Widget: nameEntry},
			{Text: "Interface Count", Widget: interfaceCount},
		},
		OnSubmit: func() {
			router := &Router{
				Name:    nameEntry.Text,
				Enabled: true,
			}
			count, _ := strconv.Atoi(interfaceCount.Text)
			router.Interfaces = make([]RouterInterface, count)
			for i := 0; i < count; i++ {
				router.Interfaces[i] = RouterInterface{
					ID:      fmt.Sprintf("eth%d", i),
					Name:    fmt.Sprintf("eth%d", i),
					Enabled: true,
				}
			}
			if err := nt.manager.CreateRouter(context.Background(), router); err != nil {
				dialog.ShowError(err, nt.window)
				return
			}
			nt.refreshRouters()
		},
	}

	dialog.ShowForm("Create Router", "Create", "Cancel", form.Items, func(confirm bool) {
		if confirm {
			form.OnSubmit()
		}
	}, nt.window)
}

func (nt *NetworkTab) showCreateTunnelDialog(protocol TunnelProtocol) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("tunnel-1")

	localIPEntry := widget.NewEntry()
	localIPEntry.SetPlaceHolder("192.168.1.1")

	remoteIPEntry := widget.NewEntry()
	remoteIPEntry.SetPlaceHolder("192.168.1.2")

	vniEntry := widget.NewEntry()
	vniEntry.SetPlaceHolder("1000")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name", Widget: nameEntry},
			{Text: "Protocol", Widget: widget.NewLabel(string(protocol))},
			{Text: "Local IP", Widget: localIPEntry},
			{Text: "Remote IP", Widget: remoteIPEntry},
			{Text: "VNI/Key", Widget: vniEntry},
		},
		OnSubmit: func() {
			vni, _ := strconv.Atoi(vniEntry.Text)
			tunnel := &Tunnel{
				Name:     nameEntry.Text,
				Protocol: protocol,
				LocalIP:  localIPEntry.Text,
				RemoteIP: remoteIPEntry.Text,
				VNI:      vni,
				Enabled:  true,
			}
			if err := nt.manager.CreateTunnel(context.Background(), tunnel); err != nil {
				dialog.ShowError(err, nt.window)
				return
			}
			nt.refreshTunnels()
		},
	}

	dialog.ShowForm(fmt.Sprintf("Create %s Tunnel", protocol), "Create", "Cancel", form.Items, func(confirm bool) {
		if confirm {
			form.OnSubmit()
		}
	}, nt.window)
}

func (nt *NetworkTab) showAddRouteDialog(routerID string) {
	destEntry := widget.NewEntry()
	destEntry.SetPlaceHolder("0.0.0.0/0")

	gatewayEntry := widget.NewEntry()
	gatewayEntry.SetPlaceHolder("10.0.0.1")

	ifaceEntry := widget.NewEntry()
	ifaceEntry.SetPlaceHolder("eth0")

	metricEntry := widget.NewEntry()
	metricEntry.SetPlaceHolder("100")
	metricEntry.SetText("100")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Destination", Widget: destEntry},
			{Text: "Gateway", Widget: gatewayEntry},
			{Text: "Interface", Widget: ifaceEntry},
			{Text: "Metric", Widget: metricEntry},
		},
		OnSubmit: func() {
			metric, _ := strconv.Atoi(metricEntry.Text)
			route := Route{
				Destination: destEntry.Text,
				Gateway:     gatewayEntry.Text,
				Interface:   ifaceEntry.Text,
				Metric:      metric,
				Type:        "static",
				Enabled:     true,
			}
			if err := nt.manager.AddRoute(context.Background(), routerID, route); err != nil {
				dialog.ShowError(err, nt.window)
				return
			}
		},
	}

	dialog.ShowForm("Add Route", "Add", "Cancel", form.Items, func(confirm bool) {
		if confirm {
			form.OnSubmit()
		}
	}, nt.window)
}

func (nt *NetworkTab) showAddNATRuleDialog(routerID string) {
	typeSelect := widget.NewSelect([]string{"snat", "dnat", "masquerade"}, nil)
	typeSelect.SetSelected("masquerade")

	sourceCIDREntry := widget.NewEntry()
	sourceCIDREntry.SetPlaceHolder("10.0.0.0/24")

	externalIPEntry := widget.NewEntry()
	externalIPEntry.SetPlaceHolder("1.2.3.4")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Type", Widget: typeSelect},
			{Text: "Source CIDR", Widget: sourceCIDREntry},
			{Text: "External IP", Widget: externalIPEntry},
		},
		OnSubmit: func() {
			// NAT rule creation logic
			dialog.ShowInformation("NAT Rule", "NAT rule creation coming soon", nt.window)
		},
	}

	dialog.ShowForm("Add NAT Rule", "Add", "Cancel", form.Items, func(confirm bool) {
		if confirm {
			form.OnSubmit()
		}
	}, nt.window)
}

func (nt *NetworkTab) showCreateBridgeDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("br0")

	dialog.ShowForm("Create Bridge", "Create", "Cancel", []*widget.FormItem{
		{Text: "Name", Widget: nameEntry},
	}, func(confirm bool) {
		if confirm {
			bridge := &Network{
				Name: nameEntry.Text,
				Type: NetworkTypeBridge,
			}
			if err := nt.manager.CreateNetwork(context.Background(), bridge); err != nil {
				dialog.ShowError(err, nt.window)
				return
			}
			nt.refreshNetworks()
		}
	}, nt.window)
}

func (nt *NetworkTab) showConnectDialog() {
	networkSelect := widget.NewSelect([]string{}, nil)
	networkNames := make([]string, len(nt.networks))
	for i, n := range nt.networks {
		networkNames[i] = n.Name
	}
	networkSelect.Options = networkNames

	routerSelect := widget.NewSelect([]string{}, nil)
	routerNames := make([]string, len(nt.routers))
	for i, r := range nt.routers {
		routerNames[i] = r.Name
	}
	routerSelect.Options = routerNames

	dialog.ShowForm("Connect Networks", "Connect", "Cancel", []*widget.FormItem{
		{Text: "Network", Widget: networkSelect},
		{Text: "Router", Widget: routerSelect},
	}, func(confirm bool) {
		if confirm {
			dialog.ShowInformation("Connect", "Network connection feature coming soon", nt.window)
		}
	}, nt.window)
}

func (nt *NetworkTab) showAssignInterfaceDialog(ifaceID string) {
	networkSelect := widget.NewSelect([]string{}, nil)
	networkNames := make([]string, len(nt.networks))
	for i, n := range nt.networks {
		networkNames[i] = n.Name
	}
	networkSelect.Options = networkNames

	dialog.ShowForm("Assign Interface", "Assign", "Cancel", []*widget.FormItem{
		{Text: "Network", Widget: networkSelect},
	}, func(confirm bool) {
		if confirm {
			dialog.ShowInformation("Assign", "Interface assignment feature coming soon", nt.window)
		}
	}, nt.window)
}

func (nt *NetworkTab) showAddVLANDialog(ifaceID string) {
	vlanEntry := widget.NewEntry()
	vlanEntry.SetPlaceHolder("100")

	dialog.ShowForm("Add VLAN", "Add", "Cancel", []*widget.FormItem{
		{Text: "VLAN ID", Widget: vlanEntry},
	}, func(confirm bool) {
		if confirm {
			if _, err := strconv.Atoi(vlanEntry.Text); err != nil {
				dialog.ShowError(fmt.Errorf("invalid VLAN ID"), nt.window)
				return
			}
			dialog.ShowInformation("VLAN", "VLAN assignment feature coming soon", nt.window)
		}
	}, nt.window)
}

func (nt *NetworkTab) showSetTrunkDialog(ifaceID string) {
	trunkEntry := widget.NewEntry()
	trunkEntry.SetPlaceHolder("100,200,300")

	dialog.ShowForm("Set Trunk VLANs", "Set", "Cancel", []*widget.FormItem{
		{Text: "VLAN IDs (comma-separated)", Widget: trunkEntry},
	}, func(confirm bool) {
		if confirm {
			dialog.ShowInformation("Trunk", "Trunk configuration feature coming soon", nt.window)
		}
	}, nt.window)
}

func (nt *NetworkTab) showRouterDetails(router *Router) {
	details := fmt.Sprintf("Router: %s\n\nID: %s\nNetworks: %d\nRoutes: %d",
		router.Name, router.ID, len(router.Interfaces), len(router.RoutingTable))
	dialog.ShowInformation("Router Details", details, nt.window)
}

// Refresh methods

func (nt *NetworkTab) refreshNetworks() {
	ctx := context.Background()
	networks, err := nt.manager.ListNetworks(ctx)
	if err != nil {
		dialog.ShowError(err, nt.window)
		return
	}
	nt.networks = networks
	nt.networkCount.Set(len(networks))
	nt.networkList.Refresh()
}

func (nt *NetworkTab) refreshRouters() {
	ctx := context.Background()
	routers, err := nt.manager.ListRouters(ctx)
	if err != nil {
		dialog.ShowError(err, nt.window)
		return
	}
	nt.routers = routers
	nt.routerCount.Set(len(routers))
	nt.routerList.Refresh()
}

func (nt *NetworkTab) refreshTunnels() {
	ctx := context.Background()
	tunnels, err := nt.manager.ListTunnels(ctx)
	if err != nil {
		dialog.ShowError(err, nt.window)
		return
	}
	nt.tunnels = tunnels
	nt.tunnelCount.Set(len(tunnels))
	nt.tunnelList.Refresh()
}

func (nt *NetworkTab) refreshInterfaces() {
	ctx := context.Background()
	interfaces, err := nt.manager.db.ListInterfaces(ctx)
	if err != nil {
		dialog.ShowError(err, nt.window)
		return
	}
	nt.interfaces = interfaces
	nt.ifaceCount.Set(len(interfaces))
	nt.ifaceList.Refresh()
}

// Delete methods

func (nt *NetworkTab) deleteNetwork(networkID string) {
	dialog.ShowConfirm("Delete Network",
		"Are you sure you want to delete this network?",
		func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := nt.manager.db.DeleteNetwork(context.Background(), networkID); err != nil {
				dialog.ShowError(err, nt.window)
				return
			}
			nt.refreshNetworks()
		}, nt.window)
}

func (nt *NetworkTab) deleteRouter(routerID string) {
	dialog.ShowConfirm("Delete Router",
		"Are you sure you want to delete this router?",
		func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := nt.manager.db.DeleteRouter(context.Background(), routerID); err != nil {
				dialog.ShowError(err, nt.window)
				return
			}
			nt.refreshRouters()
		}, nt.window)
}

func (nt *NetworkTab) deleteTunnel(tunnelID string) {
	dialog.ShowConfirm("Delete Tunnel",
		"Are you sure you want to delete this tunnel?",
		func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := nt.manager.db.DeleteTunnel(context.Background(), tunnelID); err != nil {
				dialog.ShowError(err, nt.window)
				return
			}
			nt.refreshTunnels()
		}, nt.window)
}

func (nt *NetworkTab) detachInterface(ifaceID string) {
	dialog.ShowConfirm("Detach Interface",
		"Are you sure you want to detach this interface?",
		func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := nt.manager.DetachInterface(context.Background(), ifaceID); err != nil {
				dialog.ShowError(err, nt.window)
				return
			}
			nt.refreshInterfaces()
		}, nt.window)
}

// Export/Import

func (nt *NetworkTab) exportTopology() {
	dialog.ShowInformation("Export", "Topology export feature coming soon", nt.window)
}

func (nt *NetworkTab) importTopology() {
	dialog.ShowInformation("Import", "Topology import feature coming soon", nt.window)
}

func (nt *NetworkTab) refreshFirewalls() {
	ctx := context.Background()
	firewalls, err := nt.manager.db.ListFirewalls(ctx)
	if err != nil {
		dialog.ShowError(err, nt.window)
		return
	}
	nt.firewalls = firewalls
	nt.firewallList.Refresh()
}

func (nt *NetworkTab) deleteFirewall(firewallID string) {
	dialog.ShowConfirm("Delete Firewall",
		"Are you sure you want to delete this firewall?",
		func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := nt.manager.db.DeleteFirewall(context.Background(), firewallID); err != nil {
				dialog.ShowError(err, nt.window)
				return
			}
			nt.refreshFirewalls()
		}, nt.window)
}

func (nt *NetworkTab) showCreateFirewallDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("firewall-1")

	policySelect := widget.NewSelect([]string{"accept", "drop", "reject"}, nil)
	policySelect.SetSelected("accept")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name", Widget: nameEntry},
			{Text: "Default Policy", Widget: policySelect},
		},
		OnSubmit: func() {
			firewall := &Firewall{
				Name:          nameEntry.Text,
				DefaultPolicy: policySelect.Selected,
			}
			if err := nt.manager.CreateFirewall(context.Background(), firewall); err != nil {
				dialog.ShowError(err, nt.window)
				return
			}
			nt.refreshFirewalls()
		},
	}

	dialog.ShowForm("Create Firewall", "Create", "Cancel", form.Items, func(confirm bool) {
		if confirm {
			form.OnSubmit()
		}
	}, nt.window)
}

func (nt *NetworkTab) showAddFirewallRuleDialog(firewallID string) {
	if firewallID == "" {
		dialog.ShowError(fmt.Errorf("no firewall selected"), nt.window)
		return
	}

	protocolSelect := widget.NewSelect([]string{"tcp", "udp", "icmp", "all"}, nil)
	protocolSelect.SetSelected("tcp")

	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("80")

	sourceEntry := widget.NewEntry()
	sourceEntry.SetPlaceHolder("0.0.0.0/0")

	actionSelect := widget.NewSelect([]string{"accept", "drop", "reject"}, nil)
	actionSelect.SetSelected("accept")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Protocol", Widget: protocolSelect},
			{Text: "Port", Widget: portEntry},
			{Text: "Source CIDR", Widget: sourceEntry},
			{Text: "Action", Widget: actionSelect},
		},
		OnSubmit: func() {
			port, _ := strconv.Atoi(portEntry.Text)
			rule := FirewallRule{
				Protocol:   protocolSelect.Selected,
				SourceCIDR: sourceEntry.Text,
				DestPort:   port,
				Action:     actionSelect.Selected,
			}
			if err := nt.manager.AddFirewallRule(context.Background(), firewallID, rule); err != nil {
				dialog.ShowError(err, nt.window)
				return
			}
			nt.refreshFirewalls()
		},
	}

	dialog.ShowForm("Add Firewall Rule", "Add", "Cancel", form.Items, func(confirm bool) {
		if confirm {
			form.OnSubmit()
		}
	}, nt.window)
}

// Initialize loads initial data
func (nt *NetworkTab) Initialize() {
	nt.refreshNetworks()
	nt.refreshRouters()
	nt.refreshTunnels()
	nt.refreshInterfaces()
	nt.refreshFirewalls()
	// Build topology after data is loaded
	if nt.topologyCanvas != nil {
		nt.buildTopologyView()
	}
}
