#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════╗"
echo "║    Media Ingest Server - Installation Script          ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "ERROR: Please run as root (use sudo)"
    exit 1
fi

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    echo "ERROR: Cannot detect OS"
    exit 1
fi

echo "✓ Detected OS: $OS"

# Install dependencies
echo ""
echo "Installing dependencies..."

case $OS in
    ubuntu|debian)
        apt-get update
        apt-get install -y golang-go git udev
        ;;
    fedora|centos|rhel)
        dnf install -y golang git systemd-udev
        ;;
    arch)
        pacman -Sy --noconfirm go git systemd
        ;;
    *)
        echo "WARNING: Unsupported OS. Please install Go manually."
        ;;
esac

echo "✓ Dependencies installed"

# Build the application
echo ""
echo "Building Media Ingest Server..."

if [ -f "go.mod" ]; then
    go mod download
    go build -o media-ingest ./cmd/media-ingest
    echo "✓ Build complete"
else
    echo "ERROR: go.mod not found. Are you in the correct directory?"
    exit 1
fi

# Install binary
echo ""
echo "Installing binary..."
install -m 755 media-ingest /usr/local/bin/media-ingest
echo "✓ Binary installed to /usr/local/bin/media-ingest"

# Create configuration directory
echo ""
echo "Setting up configuration..."
mkdir -p /etc/media-ingest
if [ ! -f /etc/media-ingest/config.yaml ]; then
    cp config.example.yaml /etc/media-ingest/config.yaml
    echo "✓ Configuration file created at /etc/media-ingest/config.yaml"
    echo "  Please edit this file to configure the application"
else
    echo "✓ Configuration file already exists"
fi

# Create log directory
mkdir -p /var/log/media-ingest
chmod 755 /var/log/media-ingest
echo "✓ Log directory created at /var/log/media-ingest"

# Create mount base directory
mkdir -p /mnt/ingest
chmod 755 /mnt/ingest
echo "✓ Mount directory created at /mnt/ingest"

# Create destination directory (you may want to change this)
mkdir -p /mnt/storage/media
chmod 755 /mnt/storage/media
echo "✓ Destination directory created at /mnt/storage/media"

# Install systemd service
echo ""
echo "Installing systemd service..."
if [ -f "systemd/media-ingest.service" ]; then
    cp systemd/media-ingest.service /etc/systemd/system/
    systemctl daemon-reload
    echo "✓ Systemd service installed"
else
    echo "WARNING: systemd service file not found"
fi

# Install udev rules
echo ""
echo "Installing udev rules..."
if [ -f "udev/99-media-ingest.rules" ]; then
    cp udev/99-media-ingest.rules /etc/udev/rules.d/
    udevadm control --reload-rules
    udevadm trigger
    echo "✓ Udev rules installed"
else
    echo "WARNING: udev rules file not found"
fi

# Print next steps
echo ""
echo "╔════════════════════════════════════════════════════════╗"
echo "║              Installation Complete!                    ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""
echo "Next steps:"
echo ""
echo "1. Edit the configuration file:"
echo "   sudo nano /etc/media-ingest/config.yaml"
echo ""
echo "2. Start the service:"
echo "   sudo systemctl start media-ingest"
echo ""
echo "3. Enable auto-start on boot:"
echo "   sudo systemctl enable media-ingest"
echo ""
echo "4. Check service status:"
echo "   sudo systemctl status media-ingest"
echo ""
echo "5. View logs:"
echo "   sudo journalctl -u media-ingest -f"
echo ""
echo "For manual testing, you can also run:"
echo "   sudo media-ingest -config /etc/media-ingest/config.yaml"
echo ""
