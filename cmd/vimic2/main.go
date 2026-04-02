// Vimic2 - Cluster Management & Orchestration Platform
// Main entry point
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/spf13/viper"
	"github.com/stsgym/vimic2/internal/cluster"
	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/monitor"
	"github.com/stsgym/vimic2/internal/orchestrator"
	"github.com/stsgym/vimic2/internal/ui"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

var version = "0.1.0"

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	sugar.Infow("Starting Vimic2", "version", version)

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		sugar.Warnw("Using defaults - config not found", "error", err)
		cfg = defaultConfig()
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
			hosts[name] = hypervisor.NewStubHypervisor()
		}
		hosts[name] = hv
		defer hv.Close()
	}

	// Initialize managers
	clusterMgr := cluster.NewManager(db, hosts)
	monitorMgr := monitor.NewManager(db, hosts)
	autoScaler := orchestrator.NewAutoScaler(clusterMgr, monitorMgr, sugar)

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

	// Initialize and run UI
	app := ui.NewApp(cfg, db, clusterMgr, monitorMgr, autoScaler)

	if err := app.Run(); err != nil {
		sugar.Fatalw("Application error", "error", err)
	}

	sugar.Info("Vimic2 shutdown complete")
}

// Config holds all configuration
type Config struct {
	DBPath    string
	Hosts     map[string]*hypervisor.HostConfig
	Monitor   MonitorConfig
	AutoScale AutoScaleConfig
}

type MonitorConfig struct {
	Interval   time.Duration
	Retention  time.Duration
}

type AutoScaleConfig struct {
	Enabled      bool
	CPUThreshold float64
	MemThreshold float64
	Cooldown     time.Duration
}

func loadConfig() (*Config, error) {
	// Set up Viper
	home, _ := os.UserHomeDir()
	configPath := home + "/.vimic2"

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("dbpath", configPath+"/vimic2.db")
	viper.SetDefault("monitor.interval", "5s")
	viper.SetDefault("monitor.retention", "24h")
	viper.SetDefault("autoscale.enabled", true)
	viper.SetDefault("autoscale.cpu_threshold", 70)
	viper.SetDefault("autoscale.mem_threshold", 80)
	viper.SetDefault("autoscale.cooldown", "5m")

	// Read config file
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
	}

	// Load hosts from config
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
