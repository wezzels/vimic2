// Package network provides UI components for network management
package network

import (
	"context"
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// NetworkTab provides the UI for network management
type NetworkTab struct {
	manager       *NetworkManager
	window        fyne.Window
	app           fyne.App

	// Network list
	networkList   *widget.List
	networks      []*Network
	selectedNet   *Network

	// Router list
	routerList     *widget.List
	routers        []*Router
	selectedRouter *Router

	// Tunnel list
	tunnelList    *widget.List
	tunnels       []*Tunnel
	selectedTunnel *Tunnel

	// Interface list
	ifaceList     *widget.List
	interfaces    []*VMInterface
	selectedIface *VMInterface

	// Bindings
	networkCount  binding.Int
	routerCount   binding.Int
	tunnelCount   binding.Int
	ifaceCount    binding.Int
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
		func() int { return 0 }, // TODO: Implement firewall list
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Firewall Name"),
				widget.NewLabel("Rules"),
				widget.NewLabel("Policy"),
				widget.NewLabel("Status"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			// TODO: Implement
		},
	)

	toolbar := container.NewHBox(
		widget.NewButton("Add Firewall", func() {
			nt.showCreateFirewallDialog()
		}),
		widget.NewButton("Add Rule", func() {
			nt.showAddFirewallRuleDialog("")
		}),
		widget.NewButton("Delete", func() {
			// TODO: Implement
		}),
		widget.NewButton("Refresh", func() {
			// TODO: Implement
		}),
	)

	return container.NewBorder(toolbar, nil, nil, nil, firewallList)
}

// createTopologyTab creates the network topology view
func (nt *NetworkTab) createTopologyTab() fyne.CanvasObject {
	// TODO: Implement visual topology diagram
	topologyLabel := widget.NewLabel("Network Topology\n\nDrag and drop to configure network connections")

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
		widget.NewButton("Export", func() {
			nt.exportTopology()
		}),
		widget.NewButton("Import", func() {
			nt.importTopology()
		}),
	)

	return container.NewBorder(toolbar, nil, nil, nil, topologyLabel)
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
			// TODO: Implement NAT rule creation
		},
	}

	dialog.ShowForm("Add NAT Rule", "Add", "Cancel", form.Items, func(confirm bool) {
		if confirm {
			form.OnSubmit()
		}
	}, nt.window)
}

func (nt *NetworkTab) showCreateFirewallDialog() {
	// TODO: Implement
}

func (nt *NetworkTab) showAddFirewallRuleDialog(firewallID string) {
	// TODO: Implement
}

func (nt *NetworkTab) showCreateBridgeDialog() {
	// TODO: Implement
}

func (nt *NetworkTab) showConnectDialog() {
	// TODO: Implement
}

func (nt *NetworkTab) showAssignInterfaceDialog(ifaceID string) {
	// TODO: Implement
}

func (nt *NetworkTab) showAddVLANDialog(ifaceID string) {
	// TODO: Implement
}

func (nt *NetworkTab) showSetTrunkDialog(ifaceID string) {
	// TODO: Implement
}

func (nt *NetworkTab) showRouterDetails(router *Router) {
	// TODO: Show router details in side panel
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
	// TODO: Export network topology to JSON
}

func (nt *NetworkTab) importTopology() {
	// TODO: Import network topology from JSON
}

// Initialize loads initial data
func (nt *NetworkTab) Initialize() {
	nt.refreshNetworks()
	nt.refreshRouters()
	nt.refreshTunnels()
	nt.refreshInterfaces()
}