// Vimic2 - Cluster Management & Orchestration Platform
// Main entry point
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/spf13/viper"

	"github.com/stsgym/vimic2/internal/api"
	"github.com/stsgym/vimic2/internal/cluster"
	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/monitor"
	"github.com/stsgym/vimic2/internal/orchestrator"
	"github.com/stsgym/vimic2/internal/pipeline"
	"github.com/stsgym/vimic2/internal/ui"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

var version = "0.1.0"

func main() {
	headless := flag.Bool("headless", false, "Run in headless mode (API only, no GUI)")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	sugar.Infow("Starting Vimic2", "version", version, "headless", *headless)

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		sugar.Warnw("Using defaults - config not found", "error", err)
		cfg = defaultConfig()
	}

	// Expand ~ in DB path
	if len(cfg.DBPath) > 0 && cfg.DBPath[0] == '~' {
		home, _ := os.UserHomeDir()
		cfg.DBPath = filepath.Join(home, cfg.DBPath[1:])
	}

	// Initialize database
	db, err := database.NewDB(cfg.DBPath)
	if err != nil {
		sugar.Fatalw("Failed to initialize database", "error", err)
	}
	defer db.Close()

	// Initialize hypervisors for each host
	hosts := make(map[string]hypervisor.Hypervisor)
	for name, hostCfg := range cfg.Hosts {
		hv, err := hypervisor.NewHypervisor(hostCfg)
		if err != nil {
			sugar.Warnw("Failed to connect to host", "host", name, "error", err)
			// Use stub for development
			hv = hypervisor.NewStubHypervisor()
		} else {
			defer hv.Close()
		}
		hosts[name] = hv
	}

	// Initialize managers
	clusterMgr := cluster.NewManager(db, hosts)
	monitorMgr := monitor.NewManager(db, hosts)
	autoScaler := orchestrator.NewAutoScaler(clusterMgr, monitorMgr, sugar)

	// Initialize pipeline database
	home, _ := os.UserHomeDir()
	pipelineDBPath := filepath.Join(home, ".vimic2", "pipeline.db")
	pipelineDB, err := pipeline.NewPipelineDB(pipelineDBPath)
	if err != nil {
		sugar.Warnw("Failed to initialize pipeline database — some features unavailable", "error", err)
		pipelineDB = nil
	}
	defer func() {
		if pipelineDB != nil {
			pipelineDB.Close()
		}
	}()

	// TODO: Wire up pool, network, and runner managers once interface
	// mismatches between types.PipelineDB and concrete implementations
	// are resolved. See:
	//   - PipelineDB.DeleteNetwork takes (context.Context, string) but types.PipelineDB wants (string)
	//   - PoolManager.GetPool returns (*Pool, error) but types.PoolManagerInterface wants (*PoolState, error)
	//   - NetworkManager.CreateNetwork takes (*Network) but types.NetworkManagerInterface wants (*NetworkConfig)
	//   - RunnerManager.CreateRunner takes different args than types.RunnerManagerInterface

	// Initialize artifact manager and log collector (these work with PipelineDB directly)
	var artifactMgr *pipeline.ArtifactManager
	var logCollector *pipeline.LogCollector
	if pipelineDB != nil {
		artifactMgr, err = pipeline.NewArtifactManager(pipelineDB, &pipeline.ArtifactConfig{
			StoragePath: filepath.Join(home, ".vimic2", "artifacts"),
		})
		if err != nil {
			sugar.Warnw("Failed to initialize artifact manager", "error", err)
		}
		os.MkdirAll(filepath.Join(home, ".vimic2", "artifacts"), 0755)

		logCollector, err = pipeline.NewLogCollector(pipelineDB, &pipeline.LogConfig{})
		if err != nil {
			sugar.Warnw("Failed to initialize log collector", "error", err)
		}
	}

	// Start background monitoring
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go monitorMgr.StartCollection(ctx, 5*time.Second)

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		sugar.Infow("Received shutdown signal", "signal", sig)
		cancel()
	}()

	if *headless {
		// Headless mode: start API server
		sugar.Info("Running in headless mode — starting API server")

		serverCfg := &api.ServerConfig{
			ListenAddr: ":8080",
		}
		if cfg.APIAuthEnabled {
			serverCfg.AuthEnabled = true
			serverCfg.AuthToken = cfg.APIAuthToken
		}

		server, err := api.NewServer(pipelineDB, nil, nil, artifactMgr, logCollector, nil, nil, nil, serverCfg)
		if err != nil {
			sugar.Fatalw("Failed to create API server", "error", err)
		}

		go func() {
			if err := server.Start(); err != nil {
				sugar.Errorw("API server error", "error", err)
			}
		}()

		sugar.Info("Vimic2 API server listening on :8080")
		sugar.Warn("Some features unavailable: pool/network/runner managers not yet wired (interface mismatch)")

		// Wait for shutdown signal
		<-sigChan
		server.Stop(context.Background())
	} else {
		// GUI mode: start the Fyne UI
		app := ui.NewApp(cfg, db, clusterMgr, monitorMgr, autoScaler)

		if err := app.Run(); err != nil {
			sugar.Fatalw("Application error", "error", err)
		}
	}

	sugar.Info("Vimic2 shutdown complete")
}

// Config holds all configuration
type Config struct {
	DBPath         string
	Hosts          map[string]*hypervisor.HostConfig
	Monitor        MonitorConfig
	AutoScale      AutoScaleConfig
	APIAuthEnabled bool
	APIAuthToken   string
}

type MonitorConfig struct {
	Interval  time.Duration
	Retention time.Duration
}

type AutoScaleConfig struct {
	Enabled      bool
	CPUThreshold float64
	MemThreshold float64
	Cooldown     time.Duration
}

func loadConfig() (*Config, error) {
	home, _ := os.UserHomeDir()
	configPath := home + "/.vimic2"

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")

	viper.SetDefault("dbpath", configPath+"/vimic2.db")
	viper.SetDefault("monitor.interval", "5s")
	viper.SetDefault("monitor.retention", "24h")
	viper.SetDefault("autoscale.enabled", true)
	viper.SetDefault("autoscale.cpu_threshold", 70)
	viper.SetDefault("autoscale.mem_threshold", 80)
	viper.SetDefault("autoscale.cooldown", "5m")
	viper.SetDefault("api.auth_enabled", false)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{
		DBPath: viper.GetString("dbpath"),
		Hosts:  make(map[string]*hypervisor.HostConfig),
		Monitor: MonitorConfig{
			Interval:  viper.GetDuration("monitor.interval"),
			Retention: viper.GetDuration("monitor.retention"),
		},
		AutoScale: AutoScaleConfig{
			Enabled:      viper.GetBool("autoscale.enabled"),
			CPUThreshold: viper.GetFloat64("autoscale.cpu_threshold"),
			MemThreshold: viper.GetFloat64("autoscale.mem_threshold"),
			Cooldown:     viper.GetDuration("autoscale.cooldown"),
		},
		APIAuthEnabled: viper.GetBool("api.auth_enabled"),
		APIAuthToken:   viper.GetString("api.auth_token"),
	}

	hosts := viper.GetStringMap("hosts")
	for name, h := range hosts {
		if hostMap, ok := h.(map[string]interface{}); ok {
			hostCfg := &hypervisor.HostConfig{
				Address:    getString(hostMap, "address", ""),
				Port:       getInt(hostMap, "port", 22),
				User:       getString(hostMap, "user", "root"),
				SSHKeyPath: getString(hostMap, "ssh_key_path", ""),
				Type:       getString(hostMap, "type", "libvirt"),
			}
			cfg.Hosts[name] = hostCfg
		}
	}

	return cfg, nil
}

func getString(m map[string]interface{}, key, def string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func getInt(m map[string]interface{}, key string, def int) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		}
	}
	return def
}

func defaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		DBPath: home + "/.vimic2/vimic2.db",
		Hosts:  make(map[string]*hypervisor.HostConfig),
		Monitor: MonitorConfig{
			Interval:  5 * time.Second,
			Retention: 24 * time.Hour,
		},
		AutoScale: AutoScaleConfig{
			Enabled:      true,
			CPUThreshold: 70,
			MemThreshold: 80,
			Cooldown:     5 * time.Minute,
		},
	}
}