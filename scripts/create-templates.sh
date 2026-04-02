#!/bin/bash
# Create Vimic2 CI/CD Runner Templates
# This script creates base VM templates with pre-installed tools

set -e

# Configuration
TEMPLATE_DIR="${TEMPLATE_DIR:-/var/lib/vimic2/templates}"
BASE_IMAGE="${BASE_IMAGE:-ubuntu-24.04-live-server.iso}"
BASE_IMAGE_URL="${BASE_IMAGE_URL:-https://cloud-images.ubuntu.com/releases/24.04/release/ubuntu-24.04-server-cloudimg-amd64.img}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing=()
    
    # Check for required commands
    for cmd in qemu-img virt-customize qemu-system-x86_64 wget; do
        if ! command -v $cmd &> /dev/null; then
            missing+=($cmd)
        fi
    done
    
    if [ ${#missing[@]} -gt 0 ]; then
        log_error "Missing required commands: ${missing[*]}"
        log_info "Install with: sudo apt-get install -y qemu-utils libvirt-clients wget"
        exit 1
    fi
    
    # Check for root
    if [ "$EUID" -ne 0 ]; then
        log_warn "Running without root. Some operations may fail."
    fi
    
    # Create template directory
    mkdir -p "$TEMPLATE_DIR"
    
    log_info "Prerequisites OK"
}

# Download base Ubuntu cloud image
download_base_image() {
    log_info "Downloading Ubuntu 24.04 cloud image..."
    
    local dest="$TEMPLATE_DIR/ubuntu-24.04-server-cloudimg-amd64.img"
    
    if [ -f "$dest" ]; then
        log_info "Base image already exists: $dest"
        return 0
    fi
    
    wget -O "$dest" "$BASE_IMAGE_URL"
    
    log_info "Base image downloaded: $dest"
}

# Create base template with common tools
create_base_template() {
    log_info "Creating base template..."
    
    local base="$TEMPLATE_DIR/base-ubuntu-24.04.qcow2"
    
    # Check if already exists
    if [ -f "$base" ]; then
        log_info "Base template already exists: $base"
        return 0
    fi
    
    # Create copy-on-write overlay
    qemu-img create -f qcow2 -F qcow2 -b "$TEMPLATE_DIR/ubuntu-24.04-server-cloudimg-amd64.img" "$base" 20G
    
    # Install common tools
    virt-customize -a "$base" \
        --install qemu-guest-agent,git,curl,wget,build-essential,apt-transport-https,ca-certificates,gnupg,lsb-release,software-properties-common \
        --run-command 'systemctl enable qemu-guest-agent' \
        --run-command 'systemctl start qemu-guest-agent' \
        --run-command 'curl -fsSL https://get.docker.com | sh' \
        --run-command 'systemctl enable docker' \
        --run-command 'systemctl start docker' \
        --run-command 'usermod -aG docker ubuntu' \
        --run-command 'mkdir -p /work' \
        --run-command 'chmod 777 /work'
    
    # Mark as read-only
    chmod 444 "$base"
    
    log_info "Base template created: $base"
}

# Create Go builder template
create_go_template() {
    log_info "Creating Go 1.23 template..."
    
    local dest="$TEMPLATE_DIR/base-go-1.23.qcow2"
    
    if [ -f "$dest" ]; then
        log_info "Go template already exists: $dest"
        return 0
    fi
    
    local base="$TEMPLATE_DIR/base-ubuntu-24.04.qcow2"
    
    # Create overlay from base
    qemu-img create -f qcow2 -F qcow2 -b "$base" "$dest" 30G
    
    # Install Go
    virt-customize -a "$dest" \
        --run-command 'curl -fsSL https://go.dev/dl/go1.23.0.linux-amd64.tar.gz | tar -C /usr/local -xzf -' \
        --run-command 'echo "export PATH=\$PATH:/usr/local/go/bin" >> /etc/profile' \
        --run-command 'echo "export GOPATH=/go" >> /etc/profile' \
        --run-command 'echo "export GOCACHE=/cache" >> /etc/profile' \
        --run-command 'mkdir -p /go /cache' \
        --run-command 'chmod 777 /go /cache' \
        --run-command 'go env -w GOPATH=/go' \
        --run-command 'go env -w GOCACHE=/cache' \
        --run-command 'go env -w GO111MODULE=on'
    
    # Mark as read-only
    chmod 444 "$dest"
    
    log_info "Go template created: $dest"
}

# Create Node.js template
create_node_template() {
    log_info "Creating Node 20 template..."
    
    local dest="$TEMPLATE_DIR/base-node-20.qcow2"
    
    if [ -f "$dest" ]; then
        log_info "Node template already exists: $dest"
        return 0
    fi
    
    local base="$TEMPLATE_DIR/base-ubuntu-24.04.qcow2"
    
    # Create overlay from base
    qemu-img create -f qcow2 -F qcow2 -b "$base" "$dest" 25G
    
    # Install Node.js
    virt-customize -a "$dest" \
        --run-command 'curl -fsSL https://deb.nodesource.com/setup_20.x | bash -' \
        --install nodejs \
        --run-command 'npm install -g yarn' \
        --run-command 'npm install -g playwright' \
        --run-command 'npm install -g typescript' \
        --run-command 'mkdir -p /work' \
        --run-command 'chmod 777 /work'
    
    # Mark as read-only
    chmod 444 "$dest"
    
    log_info "Node template created: $dest"
}

# Create Docker template
create_docker_template() {
    log_info "Creating Docker 27 template..."
    
    local dest="$TEMPLATE_DIR/base-docker-27.qcow2"
    
    if [ -f "$dest" ]; then
        log_info "Docker template already exists: $dest"
        return 0
    fi
    
    local base="$TEMPLATE_DIR/base-ubuntu-24.04.qcow2"
    
    # Create overlay from base
    qemu-img create -f qcow2 -F qcow2 -b "$base" "$dest" 25G
    
    # Install Docker and Kubernetes tools
    virt-customize -a "$dest" \
        --run-command 'curl -fsSL https://get.docker.com | sh' \
        --run-command 'usermod -aG docker ubuntu' \
        --run-command 'curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"' \
        --run-command 'chmod +x kubectl && mv kubectl /usr/local/bin/' \
        --run-command 'curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash' \
        --run-command 'mkdir -p /work' \
        --run-command 'chmod 777 /work'
    
    # Mark as read-only
    chmod 444 "$dest"
    
    log_info "Docker template created: $dest"
}

# Create Jenkins agent template
create_jenkins_template() {
    log_info "Creating Jenkins agent template..."
    
    local dest="$TEMPLATE_DIR/base-jenkins-agent.qcow2"
    
    if [ -f "$dest" ]; then
        log_info "Jenkins template already exists: $dest"
        return 0
    fi
    
    local base="$TEMPLATE_DIR/base-ubuntu-24.04.qcow2"
    
    # Create overlay from base
    qemu-img create -f qcow2 -F qcow2 -b "$base" "$dest" 25G
    
    # Install Java and Jenkins agent
    virt-customize -a "$dest" \
        --install openjdk-17-jre-headless,git,curl,wget \
        --run-command 'mkdir -p /home/jenkins' \
        --run-command 'useradd -m -d /home/jenkins -s /bin/bash jenkins' \
        --run-command 'mkdir -p /work' \
        --run-command 'chmod 777 /work'
    
    # Mark as read-only
    chmod 444 "$dest"
    
    log_info "Jenkins template created: $dest"
}

# List templates
list_templates() {
    log_info "Available templates:"
    echo ""
    echo "Name                    Size      Created"
    echo "----                    ----      -------"
    
    for template in "$TEMPLATE_DIR"/*.qcow2; do
        if [ -f "$template" ]; then
            local name=$(basename "$template")
            local size=$(du -h "$template" | cut -f1)
            local created=$(stat -c %y "$template" | cut -d. -f1)
            printf "%-23s %-9s %s\n" "$name" "$size" "$created"
        fi
    done
    echo ""
}

# Verify templates
verify_templates() {
    log_info "Verifying templates..."
    
    local templates=(
        "base-ubuntu-24.04.qcow2"
        "base-go-1.23.qcow2"
        "base-node-20.qcow2"
        "base-docker-27.qcow2"
    )
    
    local failed=0
    
    for template in "${templates[@]}"; do
        local path="$TEMPLATE_DIR/$template"
        
        if [ ! -f "$path" ]; then
            log_error "Missing template: $template"
            ((failed++))
            continue
        fi
        
        # Verify it's a valid qcow2
        if ! qemu-img info "$path" &> /dev/null; then
            log_error "Invalid template: $template"
            ((failed++))
            continue
        fi
        
        # Verify it has backing file (except base)
        local backing=$(qemu-img info "$path" | grep "backing file" || true)
        if [ "$template" != "base-ubuntu-24.04.qcow2" ]; then
            if [ -z "$backing" ]; then
                log_error "Template $template has no backing file"
                ((failed++))
            fi
        fi
        
        log_info "✓ $template"
    done
    
    if [ $failed -gt 0 ]; then
        log_error "$failed template(s) failed verification"
        return 1
    fi
    
    log_info "All templates verified successfully"
}

# Clean up old templates
cleanup_old_templates() {
    log_info "Cleaning up old templates..."
    
    local count=0
    
    # Find templates older than 30 days
    find "$TEMPLATE_DIR" -name "*.qcow2" -mtime +30 -type f | while read -r file; do
        log_info "Removing old template: $file"
        rm -f "$file"
        ((count++))
    done
    
    log_info "Cleaned up $count old templates"
}

# Main
main() {
    echo "=========================================="
    echo "  Vimic2 CI/CD Template Creator"
    echo "=========================================="
    echo ""
    
    case "${1:-all}" in
        "all")
            check_prerequisites
            download_base_image
            create_base_template
            create_go_template
            create_node_template
            create_docker_template
            create_jenkins_template
            list_templates
            verify_templates
            ;;
        "base")
            check_prerequisites
            download_base_image
            create_base_template
            ;;
        "go")
            check_prerequisites
            create_go_template
            ;;
        "node")
            check_prerequisites
            create_node_template
            ;;
        "docker")
            check_prerequisites
            create_docker_template
            ;;
        "jenkins")
            check_prerequisites
            create_jenkins_template
            ;;
        "list")
            list_templates
            ;;
        "verify")
            verify_templates
            ;;
        "clean")
            cleanup_old_templates
            ;;
        *)
            echo "Usage: $0 {all|base|go|node|docker|jenkins|list|verify|clean}"
            echo ""
            echo "Commands:"
            echo "  all      - Create all templates (default)"
            echo "  base     - Create base Ubuntu template only"
            echo "  go       - Create Go 1.23 template"
            echo "  node     - Create Node 20 template"
            echo "  docker   - Create Docker 27 template"
            echo "  jenkins  - Create Jenkins agent template"
            echo "  list     - List available templates"
            echo "  verify   - Verify all templates"
            echo "  clean    - Clean up old templates"
            echo ""
            echo "Environment Variables:"
            echo "  TEMPLATE_DIR  - Template directory (default: /var/lib/vimic2/templates)"
            echo "  BASE_IMAGE     - Base Ubuntu cloud image URL"
            exit 1
            ;;
    esac
    
    echo ""
    log_info "Done!"
}

main "$@"