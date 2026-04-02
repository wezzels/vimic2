#!/usr/bin/env python3
"""
Vimic2 Terminal Recording Script
Creates a terminal recording that can be played back with asciinema
"""
import subprocess
import time
import sys

# Commands to record
commands = [
    ("clear", 0.3),
    ("vimic2 --version", 0.5),
    ("vimic2 dashboard", 0.5),
    ("vimic2 cluster list", 0.5),
    ("vimic2 cluster show production-cluster", 0.5),
    ("vimic2 cluster create", 0.5),
    ("vimic2 alerts", 0.5),
    ("vimic2 node start prod-worker-1", 0.5),
    ("vimic2 status", 0.5),
]

def main():
    print("Recording Vimic2 terminal session...")
    print("=" * 60)
    print()
    print("To create a real video, run these commands in terminal:")
    print()
    print("  # 1. Install asciinema")
    print("  pip3 install asciinema")
    print()
    print("  # 2. Create recording")
    print("  asciinema rec vimic2-demo.cast")
    print()
    print("  # 3. Run these commands:")
    print("  vimic2 dashboard")
    print("  vimic2 cluster list")
    print("  vimic2 cluster show production-cluster")
    print("  vimic2 alerts")
    print("  exit")
    print()
    print("  # 4. Create GIF")
    print("  asciinema play vimic2-demo.cast")
    print("  # Or export to GIF with agg")
    print()
    print("=" * 60)
    print()
    print("\033[1;36mSimulated demo output:\033[0m")
    print()
    
    # Simulated output
    time.sleep(1)
    
    print("""
\033[1;36m╭──────────────────────────────────────────────────────────────────╮
│  ██╗  ██╗███████╗███╗   ██╗███████╗██████╗  ██████╗ ██╗███████╗   │
│  ╚██╗██╔╝██╔════╝████╗  ██║██╔════╝██╔══██╗██╔═══██╗██║██╔════╝   │
│   ╚███╔╝ █████╗  ██╔██╗ ██║█████╗  ██████╔╝██║   ██║██║███████╗   │
│   ██╔██╗ ██╔══╝  ██║╚██╗██║██╔══╝  ██╔══██╗██║   ██║██║╚════██║   │
│  ██╔╝ ██╗███████╗██║ ╚████║███████╗██║  ██║╚██████╔╝██║███████║   │
│  ╚═╝  ╚═╝╚══════╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝╚══════╝   │
│                                                                   │
│           Cluster Management Dashboard  |  v0.1.0                   │
╰──────────────────────────────────────────────────────────────────╯\033[0m
""")
    
    time.sleep(0.5)
    
    print("""
\033[1;34m┌──────────────────────────────────────────────────────────────────┐
│  VIMIC2 DASHBOARD                         [Connected] ● 10:42:15   │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ╭───────────╮  ╭───────────╮  ╭───────────╮  ╭───────────╮     │
│  │   HOSTS   │  │  CLUSTERS │  │   NODES   │  │   ALERTS  │     │
│  │\033[1;32m     3     \033[1;34m│  │\033[1;32m     5     \033[1;34m│  │\033[1;32m    18     \033[1;34m│  │\033[1;31m     2     \033[1;34m│     │
│  ╰───────────╯  ╰───────────╯  ╰───────────╯  ╰───────────╯     │
│                                                                   │
│  \033[1;36m┌─ CLUSTERS ─────────────────────────────────────────────┐\033[1;34m  │
│  │                                                            │\033[1;34m  │
│  │  ● production-cluster    3 nodes  \033[1;32m● Running\033[0m    192.168.1.0  │\033[1;34m  │
│  │  ● staging-env          5 nodes  \033[1;32m● Running\033[0m    192.168.2.0  │\033[1;34m  │
│  │  ● dev-sandbox          2 nodes  \033[1;33m◐ Deploying\033[0m  192.168.3.0  │\033[1;34m  │
│  │  ● ml-training          8 nodes  \033[1;31m● Error\033[0m     192.168.4.0  │\033[1;34m  │
│  │                                                            │\033[1;34m  │
│  \033[1;36m└────────────────────────────────────────────────────────────┘\033[1;34m  │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘\033[0m
""")
    
    time.sleep(0.5)
    
    print("""
\033[1;32m$ vimic2 cluster list\033[0m

\033[1;36m┌──────────────────────────────────────────────────────────────────┐
│  CLUSTERS                                                        │
├──────────────────────────────────────────────────────────────────┤
│  NAME                 NODES  STATUS      NETWORK          HOSTS  │
├──────────────────────────────────────────────────────────────────┤
│  production-cluster   3      \033[1;32mRunning\033[0m    192.168.1.0/24  1      │
│  staging-env          5      \033[1;32mRunning\033[0m    192.168.2.0/24  2      │
│  dev-sandbox          2      \033[1;33mDeploying\033[0m  192.168.3.0/24  1      │
│  ml-training          8      \033[1;31mError\033[0m      192.168.4.0/24  3      │
│  test-bed             2      \033[1;30mStopped\033[0m    192.168.5.0/24  1      │
└──────────────────────────────────────────────────────────────────┘\033[0m
""")
    
    time.sleep(0.5)
    
    print("""
\033[1;32m$ vimic2 cluster show production-cluster\033[0m

\033[1;36m┌──────────────────────────────────────────────────────────────────┐
│  CLUSTER: production-cluster                    \033[1;32m● Running\033[0m        │
├──────────────────────────────────────────────────────────────────┤
│  NODES                                                             │
│  ──────────────────────────────────────────────────────────────── │
│  NODE              ROLE    STATUS    IP              CPU    MEM    │
│  prod-master-1    master  \033[1;32m●\033[0m       192.168.1.10   23%   45%    │
│  prod-worker-1    worker  \033[1;32m●\033[0m       192.168.1.11   67%   72%    │
│  prod-worker-2    worker  \033[1;32m●\033[0m       192.168.1.12   45%   58%    │
│                                                                   │
│  AUTO-SCALING: Enabled  |  Min: 2  |  Max: 10  |  Current: 3    │
└──────────────────────────────────────────────────────────────────┘\033[0m
""")
    
    time.sleep(0.5)
    
    print("""
\033[1;32m$ vimic2 alerts\033[0m

\033[1;36m┌──────────────────────────────────────────────────────────────────┐
│  ALERTS                                          Active: 2        │
├──────────────────────────────────────────────────────────────────┤
│  \033[1;31mCRITICAL\033[0m                                                        │
│  ──────────────────────────────────────────────────────────────── │
│  ● High CPU on ml-gpu-3 (95%)                                    │
│    Fired: 10:43:01  |  Duration: 5m 14s  |  [\033[1;32mAcknowledge\033[0m]         │
│                                                                   │
│  \033[1;33mWARNING\033[0m                                                         │
│  ──────────────────────────────────────────────────────────────── │
│  ● High Memory on prod-worker-2 (82%)                            │
│    Fired: 10:41:22  |  Duration: 8m 53s  |  [\033[1;32mAcknowledge\033[0m]         │
│                                                                   │
│  \033[1;32mRESOLVED\033[0m                                                        │
│  ──────────────────────────────────────────────────────────────── │
│  ✓ Disk Full on dev-sandbox            [Resolved 10:32]         │
│  ✓ Node Down on test-1               [Resolved 10:28]         │
└──────────────────────────────────────────────────────────────────┘\033[0m
""")
    
    time.sleep(0.5)
    
    print("""
\033[1;32m$ vimic2 node start prod-worker-1\033[0m

Starting node prod-worker-1...
\033[1;32m✓ Node prod-worker-1 started successfully\033[0m
\033[1;36m  IP: 192.168.1.11
  Status: Running
  Uptime: 0h 2m\033[0m
""")
    
    print("\n\033[1;32mDemo complete! 🎬\033[0m\n")

if __name__ == "__main__":
    main()
