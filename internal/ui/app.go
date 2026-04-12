// Package ui provides the Fyne UI implementation for Vimic2
package ui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/stsgym/vimic2/internal/cluster"
	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/deploy"
	"github.com/stsgym/vimic2/internal/monitor"
	"github.com/stsgym/vimic2/internal/orchestrator"
	"github.com/stsgym/vimic2/internal/status"
)

type App struct {
	window          fyne.Window
	db              *database.DB
	clusterMgr      *cluster.Manager
	monitorMgr      *monitor.Manager
	autoScaler      *orchestrator.AutoScaler
	statusWatcher   *status.Watcher
	deployWizard    *deploy.Wizard
	clusters        []*database.Cluster
	selectedCluster *database.Cluster
	selectedNode    *database.Node
}

func NewApp(
	cfg interface{},
	db *database.DB,
	clusterMgr *cluster.Manager,
	monitorMgr *monitor.Manager,
	autoScaler *orchestrator.AutoScaler,
) *App {
	app := &App{
		db:         db,
		clusterMgr: clusterMgr,
		monitorMgr: monitorMgr,
		autoScaler: autoScaler,
	}

	// Create status watcher
	app.statusWatcher = status.NewWatcher(db, nil) // hosts will be added later

	return app
}

func (a *App) Run() error {
	a.window = a.makeMainWindow()
	a.window.Resize(fyne.NewSize(1200, 800))

	// Start status watching
	a.statusWatcher.Start(5 * time.Second)

	a.window.ShowAndRun()
	return nil
}

// ============== Main Window ==============

func (a *App) makeMainWindow() fyne.Window {
	app := fyne.CurrentApp()
	win := app.NewWindow("Vimic2 - Cluster Management")

	toolbar := a.makeToolbar()

	split := container.NewHSplit(
		a.makeSidebar(),
		a.makeMainContent(),
	)
	split.SetOffset(0.25)

	win.SetContent(container.NewBorder(toolbar, nil, nil, nil, split))
	win.SetMainMenu(a.makeMenu())

	return win
}

func (a *App) makeMenu() *fyne.MainMenu {
	return fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("New Cluster", a.onNewCluster),
			fyne.NewMenuItem("Add Host", a.onAddHost),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Refresh", a.onRefresh),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Settings", a.onSettings),
			fyne.NewMenuItem("Quit", func() { fyne.CurrentApp().Quit() }),
		),
		fyne.NewMenu("Cluster",
			fyne.NewMenuItem("Deploy", a.onDeployCluster),
			fyne.NewMenuItem("Scale", a.onScaleCluster),
			fyne.NewMenuItem("Delete", a.onDeleteCluster),
		),
		fyne.NewMenu("Node",
			fyne.NewMenuItem("Start", a.onStartNode),
			fyne.NewMenuItem("Stop", a.onStopNode),
			fyne.NewMenuItem("Restart", a.onRestartNode),
			fyne.NewMenuItem("Console", a.onNodeConsole),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Details", a.onNodeDetails),
			fyne.NewMenuItem("Delete", a.onDeleteNode),
		),
		fyne.NewMenu("Help",
			fyne.NewMenuItem("Documentation", nil),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("About", a.onAbout),
		),
	)
}

func (a *App) makeToolbar() fyne.CanvasObject {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.ViewRefreshIcon(), a.onRefresh),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ContentAddIcon(), a.onNewCluster),
		widget.NewToolbarAction(theme.FolderOpenIcon(), a.onAddHost),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.MediaPlayIcon(), a.onDeployCluster),
		widget.NewToolbarAction(theme.DeleteIcon(), a.onDeleteCluster),
	)
}

// ============== Sidebar ==============

func (a *App) makeSidebar() fyne.CanvasObject {
	clusters, _ := a.clusterMgr.ListClusters()
	hosts, _ := a.db.ListHosts()
	a.clusters = clusters

	header := widget.NewLabel("Vimic2")
	header.TextStyle = fyne.TextStyle{Bold: true}

	// Hosts section
	hostsHeader := widget.NewLabel("Hosts")
	hostsHeader.TextStyle = fyne.TextStyle{Bold: true}

	hostsList := container.NewVBox()
	for _, h := range hosts {
		btn := widget.NewButton(h.Name, func() {})
		btn.Importance = widget.LowImportance
		hostsList.Add(btn)
	}
	if len(hosts) == 0 {
		hostsList.Add(widget.NewLabel("No hosts"))
	}

	// Clusters section
	clustersHeader := widget.NewLabel("Clusters")
	clustersHeader.TextStyle = fyne.TextStyle{Bold: true}

	clustersBox := container.NewVBox()
	for _, c := range clusters {
		status := "🟡"
		switch c.Status {
		case "running":
			status = "🟢"
		case "error":
			status = "🔴"
		case "deploying":
			status = "🔵"
		}
		btn := widget.NewButton(fmt.Sprintf("%s %s", status, c.Name), func() {
			a.selectedCluster = c
			a.selectedNode = nil
			a.refreshMainContent()
		})
		btn.Importance = widget.LowImportance
		clustersBox.Add(btn)
	}
	if len(clusters) == 0 {
		clustersBox.Add(widget.NewLabel("No clusters"))
	}

	newClusterBtn := widget.NewButton("+ New Cluster", a.onNewCluster)
	newClusterBtn.Importance = widget.HighImportance

	return container.NewVBox(
		header,
		widget.NewSeparator(),
		hostsHeader,
		hostsList,
		widget.NewSeparator(),
		clustersHeader,
		container.NewVScroll(clustersBox),
		widget.NewSeparator(),
		newClusterBtn,
	)
}

// ============== Main Content ==============

func (a *App) makeMainContent() fyne.CanvasObject {
	if a.selectedCluster != nil {
		if a.selectedNode != nil {
			return a.makeNodeDetail()
		}
		return a.makeClusterDetail()
	}
	return a.makeDashboard()
}

func (a *App) makeDashboard() fyne.CanvasObject {
	clusters, _ := a.clusterMgr.ListClusters()
	hosts, _ := a.db.ListHosts()

	statsRow := container.NewGridWithColumns(4,
		a.makeStatCard("Hosts", fmt.Sprintf("%d", len(hosts)), theme.ComputerIcon()),
		a.makeStatCard("Clusters", fmt.Sprintf("%d", len(clusters)), theme.FolderIcon()),
		a.makeStatCard("Nodes", a.countTotalNodes(clusters), theme.DocumentIcon()),
		a.makeStatCard("Running", a.countRunningNodes(clusters), theme.ConfirmIcon()),
	)

	actions := container.NewGridWithColumns(4,
		widget.NewButton("New Cluster", a.onNewCluster),
		widget.NewButton("Add Host", a.onAddHost),
		widget.NewButton("Refresh", a.onRefresh),
		widget.NewButton("Settings", a.onSettings),
	)

	recentHeader := widget.NewLabel("Recent Clusters")
	recentHeader.TextStyle = fyne.TextStyle{Bold: true}

	recentList := container.NewVBox()
	for i, c := range clusters {
		if i >= 5 {
			break
		}
		row := container.NewGridWithColumns(3,
			widget.NewLabel(c.Name),
			widget.NewLabel(c.Status),
			widget.NewButton("View", func() {
				a.selectedCluster = c
				a.refreshMainContent()
			}),
		)
		recentList.Add(row)
	}

	return container.NewVBox(
		widget.NewLabel("Dashboard"),
		widget.NewSeparator(),
		statsRow,
		widget.NewSeparator(),
		actions,
		widget.NewSeparator(),
		recentHeader,
		recentList,
		layout.NewSpacer(),
	)
}

func (a *App) makeStatCard(title, value string, icon fyne.Resource) fyne.CanvasObject {
	iconWidget := widget.NewIcon(icon)
	return container.NewVBox(
		iconWidget,
		widget.NewLabelWithStyle(value, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{}),
	)
}

// ============== Cluster Detail ==============

func (a *App) makeClusterDetail() fyne.CanvasObject {
	c := a.selectedCluster
	nodes, _ := a.db.ListClusterNodes(c.ID)

	total, running, stopped := a.countNodeStates(nodes)

	header := container.NewGridWithColumns(2,
		widget.NewLabel(c.Name),
		container.NewHBox(
			widget.NewButton("Deploy", a.onDeployCluster),
			widget.NewButton("Scale", a.onScaleCluster),
			widget.NewButton("Delete", a.onDeleteCluster),
			widget.NewButton("← Back", func() {
				a.selectedCluster = nil
				a.refreshMainContent()
			}),
		),
	)

	statusCard := container.NewGridWithColumns(4,
		a.makeStatCard("Total", fmt.Sprintf("%d", total), theme.ComputerIcon()),
		a.makeStatCard("Running", fmt.Sprintf("%d", running), theme.ConfirmIcon()),
		a.makeStatCard("Stopped", fmt.Sprintf("%d", stopped), theme.CancelIcon()),
		a.makeStatCard("Errors", "0", theme.WarningIcon()),
	)

	tableHeader := container.NewGridWithColumns(6,
		widget.NewLabel("Name"),
		widget.NewLabel("Role"),
		widget.NewLabel("State"),
		widget.NewLabel("IP"),
		widget.NewLabel("CPU"),
		widget.NewLabel("Actions"),
	)

	nodesList := container.NewVBox()
	for _, n := range nodes {
		stateIcon := theme.CancelIcon()
		if n.State == "running" {
			stateIcon = theme.ConfirmIcon()
		}

		// Get latest metric for this node
		cpuText := "-"
		if metric, _ := a.db.GetLatestMetric(n.ID); metric != nil {
			cpuText = fmt.Sprintf("%.1f%%", metric.CPU)
		}

		row := container.NewGridWithColumns(6,
			widget.NewButton(n.Name, func() {
				a.selectedNode = n
				a.refreshMainContent()
			}),
			widget.NewLabel(n.Role),
			container.NewHBox(widget.NewIcon(stateIcon), widget.NewLabel(n.State)),
			widget.NewLabel(n.IP),
			widget.NewLabel(cpuText),
			container.NewHBox(
				widget.NewButton("▶", func() { a.onNodeAction(n, "start") }),
				widget.NewButton("■", func() { a.onNodeAction(n, "stop") }),
				widget.NewButton("↻", func() { a.onNodeAction(n, "restart") }),
				widget.NewButton("🗑", func() { a.onNodeDelete(n) }),
			),
		)
		nodesList.Add(row)
	}

	nodesScroll := container.NewVScroll(nodesList)

	addNodeBtn := widget.NewButton("+ Add Node", func() {
		a.selectedNode = &database.Node{ClusterID: c.ID}
		a.onNewNode()
	})

	return container.NewVBox(
		header,
		widget.NewSeparator(),
		statusCard,
		widget.NewSeparator(),
		container.NewHBox(widget.NewLabel("Nodes"), addNodeBtn),
		tableHeader,
		widget.NewSeparator(),
		nodesScroll,
	)
}

// ============== Node Detail ==============

func (a *App) makeNodeDetail() fyne.CanvasObject {
	n := a.selectedNode

	// Get node metrics history (last hour)
	metrics, _ := a.db.GetNodeMetrics(n.ID, time.Now().Add(-time.Hour))

	header := container.NewGridWithColumns(2,
		container.NewVBox(
			widget.NewLabel(n.Name),
			widget.NewLabel(fmt.Sprintf("Role: %s | Host: %s", n.Role, n.HostID)),
		),
		container.NewHBox(
			widget.NewButton("▶ Start", func() { a.onNodeAction(n, "start") }),
			widget.NewButton("■ Stop", func() { a.onNodeAction(n, "stop") }),
			widget.NewButton("↻ Restart", func() { a.onNodeAction(n, "restart") }),
			widget.NewButton("← Back", func() {
				a.selectedNode = nil
				a.refreshMainContent()
			}),
		),
	)

	// State card
	stateCard := container.NewVBox(
		widget.NewLabel("Status"),
		widget.NewLabel(n.State),
	)

	// IP card
	ipCard := container.NewVBox(
		widget.NewLabel("IP Address"),
		widget.NewLabel(n.IP),
	)

	// Metrics cards
	cpuCard := container.NewVBox(
		widget.NewLabel("CPU"),
		widget.NewLabel("-"),
	)
	memCard := container.NewVBox(
		widget.NewLabel("Memory"),
		widget.NewLabel("-"),
	)
	diskCard := container.NewVBox(
		widget.NewLabel("Disk"),
		widget.NewLabel("-"),
	)

	// Update with latest if available
	if len(metrics) > 0 {
		latest := metrics[len(metrics)-1]
		cpuCard = container.NewVBox(
			widget.NewLabel("CPU"),
			widget.NewLabel(fmt.Sprintf("%.1f%%", latest.CPU)),
		)
		memCard = container.NewVBox(
			widget.NewLabel("Memory"),
			widget.NewLabel(fmt.Sprintf("%.1f%%", latest.Memory)),
		)
		diskCard = container.NewVBox(
			widget.NewLabel("Disk"),
			widget.NewLabel(fmt.Sprintf("%.1f%%", latest.Disk)),
		)
	}

	// Metrics chart (simple representation)
	chartLabel := widget.NewLabel("Metrics History (last hour)")
	chartLabel.TextStyle = fyne.TextStyle{Bold: true}

	metricsBox := container.NewVBox()
	for _, m := range metrics {
		timeStr := m.RecordedAt.Format("15:04:05")
		row := widget.NewLabel(fmt.Sprintf("%s | CPU: %.1f%% | Mem: %.1f%% | Disk: %.1f%%",
			timeStr, m.CPU, m.Memory, m.Disk))
		metricsBox.Add(row)
	}
	if len(metrics) == 0 {
		metricsBox.Add(widget.NewLabel("No metrics data available"))
	}
	metricsScroll := container.NewVScroll(metricsBox)

	return container.NewVBox(
		header,
		widget.NewSeparator(),
		container.NewGridWithColumns(4,
			stateCard, ipCard, cpuCard, memCard,
		),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			diskCard,
			container.NewVBox(
				widget.NewLabel("Uptime"),
				widget.NewLabel("-"),
			),
		),
		widget.NewSeparator(),
		chartLabel,
		widget.NewSeparator(),
		metricsScroll,
	)
}

func (a *App) refreshMainContent() {
	content := a.makeMainContent()
	a.window.SetContent(container.NewVBox(content))
}

// ============== Helpers ==============

func (a *App) countTotalNodes(clusters []*database.Cluster) string {
	total := 0
	for _, c := range clusters {
		nodes, _ := a.db.ListClusterNodes(c.ID)
		total += len(nodes)
	}
	return fmt.Sprintf("%d", total)
}

func (a *App) countRunningNodes(clusters []*database.Cluster) string {
	total := 0
	for _, c := range clusters {
		nodes, _ := a.db.ListClusterNodes(c.ID)
		for _, n := range nodes {
			if n.State == "running" {
				total++
			}
		}
	}
	return fmt.Sprintf("%d", total)
}

func (a *App) countNodeStates(nodes []*database.Node) (total, running, stopped int) {
	for _, n := range nodes {
		total++
		if n.State == "running" {
			running++
		} else {
			stopped++
		}
	}
	return
}

// ============== Dialogs ==============

func (a *App) onNewCluster() {
	a.deployWizard = deploy.NewWizard()
	a.showDeployWizard()
}

func (a *App) showDeployWizard() {
	if a.deployWizard == nil {
		a.deployWizard = deploy.NewWizard()
	}

	step := a.deployWizard.GetStep()

	switch step {
	case 1:
		a.showWizardStep1()
	case 2:
		a.showWizardStep2()
	case 3:
		a.showWizardStep3()
	case 4:
		a.showWizardStep4()
	}
}

func (a *App) showWizardStep1() {
	cluster := a.deployWizard.GetCluster()
	nameEntry := widget.NewEntry()
	nameEntry.SetText(cluster.Name)
	nameEntry.SetPlaceHolder("my-cluster")

	dialog.NewForm("Step 1: Cluster Name", "Next", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Name", nameEntry),
		},
		func(confirmed bool) {
			if confirmed {
				a.deployWizard.SetName(nameEntry.Text)
				a.deployWizard.NextStep()
				a.showDeployWizard()
			}
		}, a.window).Show()
}

func (a *App) showWizardStep2() {
	// Show preset templates
	templates := []string{"dev", "prod", "db"}
	presetEntry := widget.NewSelectEntry(templates)

	dialog.NewForm("Step 2: Choose Template", "Next", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Preset", presetEntry),
		},
		func(confirmed bool) {
			if confirmed {
				preset := presetEntry.Text
				if preset != "" {
					if t := deploy.GetPreset(preset); t != nil {
						for _, ng := range t.NodeGroups {
							a.deployWizard.AddNodeGroup(ng)
						}
					}
				}
				a.deployWizard.NextStep()
				a.showDeployWizard()
			}
		}, a.window).Show()
}

func (a *App) showWizardStep3() {
	dialog.ShowInformation("Step 3", "Customize nodes if needed", a.window)
	a.deployWizard.NextStep()
	a.showDeployWizard()
}

func (a *App) showWizardStep4() {
	cluster := a.deployWizard.GetCluster()
	summary := fmt.Sprintf("Cluster: %s\nNodes: %d",
		cluster.Name, len(cluster.NodeGroups))

	dialog.NewForm("Step 4: Review & Deploy", "Deploy", "Cancel",
		[]*widget.FormItem{},
		func(confirmed bool) {
			if confirmed {
				if err := a.deployWizard.Validate(); err != nil {
					dialog.ShowError(err, a.window)
					return
				}
				// Execute deployment
				cluster := a.deployWizard.GetCluster()
				go func() {
					ctx := context.Background()
					if err := a.clusterMgr.DeployCluster(ctx, cluster.ID); err != nil {
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "Deployment Failed",
							Content: err.Error(),
						})
					} else {
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "Deployment Complete",
							Content: fmt.Sprintf("Cluster %s deployed successfully", cluster.Name),
						})
					}
				}()
				a.refreshMainContent()
			}
		}, a.window).Show()

	dialog.ShowInformation("Review", summary, a.window)
}

func (a *App) onAddHost() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("my-server")
	addrEntry := widget.NewEntry()
	addrEntry.SetPlaceHolder("192.168.1.100")
	userEntry := widget.NewEntry()
	userEntry.SetText("root")

	dialog.NewForm("Add Host", "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Name", nameEntry),
			widget.NewFormItem("Address", addrEntry),
			widget.NewFormItem("SSH User", userEntry),
		},
		func(confirmed bool) {
			if !confirmed {
				return
			}
			host := &database.Host{
				ID:      nameEntry.Text,
				Name:    nameEntry.Text,
				Address: addrEntry.Text,
				User:    userEntry.Text,
				Port:    22,
				HVType:  "libvirt",
			}
			a.db.SaveHost(host)
			a.refreshMainContent()
		}, a.window).Show()
}

func (a *App) onScaleCluster() {
	if a.selectedCluster == nil {
		return
	}
	countEntry := widget.NewEntry()
	countEntry.SetPlaceHolder("3")

	dialog.NewForm("Scale Cluster", "Scale", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Desired Nodes", countEntry),
		},
		func(confirmed bool) {
			if !confirmed {
				return
			}
			// Parse count and scale cluster
			var count int
			if _, err := fmt.Sscanf(countEntry.Text, "%d", &count); err == nil {
				go func() {
					ctx := context.Background()
					if err := a.clusterMgr.ScaleCluster(ctx, a.selectedCluster.ID, count); err != nil {
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "Scale Failed",
							Content: err.Error(),
						})
					}
				}()
			}
		}, a.window).Show()
}

func (a *App) onNodeAction(n *database.Node, action string) {
	fmt.Printf("Node action: %s on %s\n", action, n.Name)
	// Execute action via cluster manager
	go func() {
		ctx := context.Background()
		var err error
		switch action {
		case "start":
			err = a.clusterMgr.StartNode(ctx, n.ID)
		case "stop":
			err = a.clusterMgr.StopNode(ctx, n.ID)
		case "restart":
			err = a.clusterMgr.RestartNode(ctx, n.ID)
		}
		if err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Action Failed",
				Content: fmt.Sprintf("Failed to %s %s: %v", action, n.Name, err),
			})
		}
	}()
	a.refreshMainContent()
}

func (a *App) onNodeDelete(n *database.Node) {
	dialog.ShowConfirm("Delete Node", fmt.Sprintf("Delete node %s?", n.Name),
		func(confirmed bool) {
			if confirmed {
				a.refreshMainContent()
			}
		}, a.window)
}

func (a *App) onSettings() {
	dialog.ShowInformation("Settings", "Settings panel coming soon", a.window)
}

func (a *App) onAbout() {
	dialog.ShowInformation("About Vimic2",
		"Vimic2 v0.1.0\n\nCluster Management Platform\n\nPure Go + Fyne", a.window)
}

// ============== Menu Handlers ==============

func (a *App) onRefresh() {
	a.clusters, _ = a.clusterMgr.ListClusters()
	a.refreshMainContent()
}

func (a *App) onDeployCluster() {
	if a.selectedCluster == nil {
		return
	}
	dialog.ShowInformation("Deploy", fmt.Sprintf("Deploying %s...", a.selectedCluster.Name), a.window)
}

func (a *App) onDeleteCluster() {
	if a.selectedCluster == nil {
		return
	}
	dialog.ShowConfirm("Delete", fmt.Sprintf("Delete %s?", a.selectedCluster.Name),
		func(confirmed bool) {
			if confirmed {
				a.selectedCluster = nil
				a.refreshMainContent()
			}
		}, a.window)
}

func (a *App) onNewNode() {
	nameEntry := widget.NewEntry()
	roleEntry := widget.NewSelectEntry([]string{"worker", "database", "master"})
	roleEntry.SetText("worker")

	dialog.NewForm("Add Node", "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Name", nameEntry),
			widget.NewFormItem("Role", roleEntry),
		},
		func(confirmed bool) {
			if !confirmed {
				return
			}
			a.refreshMainContent()
		}, a.window).Show()
}

func (a *App) onStartNode()   { a.onNodeAction(a.selectedNode, "start") }
func (a *App) onStopNode()    { a.onNodeAction(a.selectedNode, "stop") }
func (a *App) onRestartNode() { a.onNodeAction(a.selectedNode, "restart") }
func (a *App) onDeleteNode()  { a.onNodeDelete(a.selectedNode) }
func (a *App) onNodeConsole() { dialog.ShowInformation("Console", "VNC/SPICE coming soon", a.window) }
func (a *App) onNodeDetails() {
	if a.selectedNode != nil {
		a.refreshMainContent()
	}
}
