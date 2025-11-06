# Deployment Guide

## System Requirements

### Hardware
- **CPU**: 2+ cores recommended for concurrent transfers
- **RAM**: 2GB minimum, 4GB+ recommended
- **Storage**: Sufficient space for media files
- **I/O**: Fast storage recommended (SSD) for destination

### Software
- **OS**: Linux (Ubuntu 20.04+, Debian 11+, CentOS 8+, Fedora 35+, Arch)
- **Kernel**: 5.x or higher (for better device detection)
- **Go**: 1.21 or higher
- **systemd**: For service management
- **udev**: For device detection

## Pre-Installation Checklist

- [ ] Root/sudo access available
- [ ] Go installed and in PATH
- [ ] Sufficient disk space for destination storage
- [ ] Network configured (for email notifications)
- [ ] Backup existing data if upgrading

## Installation Methods

### Method 1: Automated Installation (Recommended)

```bash
# Clone or download the repository
cd AutoFileIngest

# Make install script executable
chmod +x install.sh

# Run installation
sudo ./install.sh
```

### Method 2: Using Makefile

```bash
# Build and install
make build
sudo make install

# Or combine both steps
sudo make install
```

### Method 3: Manual Installation

See README.md for detailed manual installation steps.

## Post-Installation Configuration

### 1. Edit Configuration File

```bash
sudo nano /etc/media-ingest/config.yaml
```

**Critical settings to review:**

```yaml
# Set your destination path
destination_path: "/mnt/storage/media"  # Change this!

# Adjust worker count based on CPU
transfer:
  max_workers: 4  # Set to number of CPU cores

# Configure email if needed
email:
  enabled: true  # Set to true if you want notifications
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "your-email@gmail.com"
  password: "your-app-password"  # Use app-specific password!
  to:
    - "admin@example.com"
```

### 2. Create Destination Directory

```bash
# Create and set permissions
sudo mkdir -p /mnt/storage/media
sudo chmod 755 /mnt/storage/media

# Optional: Set specific owner
sudo chown mediauser:mediagroup /mnt/storage/media
```

### 3. Configure Permissions

The service runs as root by default (required for mounting). To run as a different user:

```bash
# Edit systemd service
sudo nano /etc/systemd/system/media-ingest.service

# Change User and Group
[Service]
User=mediauser
Group=mediagroup

# Reload systemd
sudo systemctl daemon-reload
```

### 4. Test Configuration

```bash
# Validate config file
sudo media-ingest -config /etc/media-ingest/config.yaml &
# Check for errors, then stop with Ctrl+C
```

## Starting the Service

```bash
# Start service
sudo systemctl start media-ingest

# Check status
sudo systemctl status media-ingest

# Enable auto-start on boot
sudo systemctl enable media-ingest

# View live logs
sudo journalctl -u media-ingest -f
```

## Monitoring

### Service Status

```bash
# Check if service is running
sudo systemctl is-active media-ingest

# View detailed status
sudo systemctl status media-ingest
```

### Logs

```bash
# Live log viewing
sudo journalctl -u media-ingest -f

# Last 100 lines
sudo journalctl -u media-ingest -n 100

# Logs from specific time
sudo journalctl -u media-ingest --since "1 hour ago"

# Logs from today
sudo journalctl -u media-ingest --since today
```

### Server Logs

```bash
# List all server logs
ls -lh /var/log/media-ingest/

# View latest log
tail -f /var/log/media-ingest/server_*.log

# Search logs for errors
grep ERROR /var/log/media-ingest/*.log
```

## Firewall Configuration

If using email notifications:

```bash
# UFW (Ubuntu/Debian)
sudo ufw allow out 587/tcp  # SMTP
sudo ufw allow out 465/tcp  # SMTPS (if using SSL)

# firewalld (CentOS/Fedora)
sudo firewall-cmd --permanent --add-port=587/tcp
sudo firewall-cmd --reload
```

## Performance Optimization

### For High-Volume Transfers

Edit `/etc/media-ingest/config.yaml`:

```yaml
transfer:
  max_workers: 8  # Increase based on CPU cores
  buffer_size: 4194304  # 4MB buffer for large files
  verify_checksums: true  # Keep enabled for safety
```

### For Resource-Constrained Systems

```yaml
transfer:
  max_workers: 2  # Reduce worker count
  buffer_size: 524288  # 512KB buffer
```

### Kernel Tuning (Advanced)

```bash
# Increase inotify limits
sudo sysctl -w fs.inotify.max_user_watches=524288
sudo sysctl -w fs.inotify.max_queued_events=32768

# Make permanent
echo "fs.inotify.max_user_watches=524288" | sudo tee -a /etc/sysctl.conf
echo "fs.inotify.max_queued_events=32768" | sudo tee -a /etc/sysctl.conf
```

## Backup and Recovery

### Backup Configuration

```bash
# Backup config and logs
sudo tar -czf media-ingest-backup-$(date +%Y%m%d).tar.gz \
  /etc/media-ingest \
  /var/log/media-ingest

# Store backup safely
sudo mv media-ingest-backup-*.tar.gz /backup/location/
```

### Restore Configuration

```bash
# Extract backup
sudo tar -xzf media-ingest-backup-YYYYMMDD.tar.gz -C /

# Restart service
sudo systemctl restart media-ingest
```

## Upgrading

```bash
# Stop service
sudo systemctl stop media-ingest

# Backup current installation
sudo cp /usr/local/bin/media-ingest /usr/local/bin/media-ingest.backup
sudo cp /etc/media-ingest/config.yaml /etc/media-ingest/config.yaml.backup

# Pull latest code
git pull

# Rebuild and reinstall
make build
sudo make install

# Compare configs
diff /etc/media-ingest/config.yaml.backup /etc/media-ingest/config.yaml

# Restart service
sudo systemctl start media-ingest

# Check status
sudo systemctl status media-ingest
```

## Troubleshooting

### Service Won't Start

```bash
# Check service status for errors
sudo systemctl status media-ingest

# Check logs
sudo journalctl -u media-ingest -n 50 --no-pager

# Verify binary exists
ls -la /usr/local/bin/media-ingest

# Verify config is valid
sudo media-ingest -config /etc/media-ingest/config.yaml
```

### Devices Not Detected

```bash
# Check udev rules
ls -la /etc/udev/rules.d/99-media-ingest.rules

# Reload udev
sudo udevadm control --reload-rules
sudo udevadm trigger

# Monitor udev events
sudo udevadm monitor --environment --udev

# Check device detection
lsblk
sudo blkid
```

### Mounting Failures

```bash
# Check mount directory permissions
ls -la /mnt/ingest/

# Check mount point availability
df -h | grep ingest

# Manual mount test
sudo mkdir -p /mnt/ingest/test
sudo mount /dev/sdb1 /mnt/ingest/test

# Check dmesg for errors
sudo dmesg | tail -20
```

### Permission Errors

```bash
# Check destination permissions
ls -la /mnt/storage/media/

# Fix permissions
sudo chmod -R 755 /mnt/storage/media/
sudo chown -R root:root /mnt/storage/media/

# Check log directory permissions
sudo chmod 755 /var/log/media-ingest
```

### Checksum Failures

Common causes:
- Corrupted source media
- Bad cable/connection
- Disk errors
- Insufficient disk space

```bash
# Check disk health
sudo smartctl -a /dev/sdb  # For source device
sudo smartctl -a /dev/sda  # For destination

# Check disk space
df -h

# Manual checksum test
sha256sum /path/to/source/file
sha256sum /path/to/dest/file
```

### High CPU Usage

```bash
# Check worker count
grep max_workers /etc/media-ingest/config.yaml

# Reduce workers if needed
sudo nano /etc/media-ingest/config.yaml
# Set max_workers to lower value

# Restart service
sudo systemctl restart media-ingest
```

### Memory Issues

```bash
# Check memory usage
free -h
top -o %MEM

# Reduce buffer size
sudo nano /etc/media-ingest/config.yaml
# Set buffer_size to smaller value (e.g., 524288)

# Restart service
sudo systemctl restart media-ingest
```

## Security Hardening

### Run as Non-Root User (Advanced)

```bash
# Create dedicated user
sudo useradd -r -s /bin/false media-ingest

# Add to disk group
sudo usermod -a -G disk media-ingest

# Update systemd service
sudo nano /etc/systemd/system/media-ingest.service
# Change User=root to User=media-ingest

# Configure sudo for mounting
sudo visudo
# Add: media-ingest ALL=(ALL) NOPASSWD: /bin/mount, /bin/umount

# Reload and restart
sudo systemctl daemon-reload
sudo systemctl restart media-ingest
```

### Encrypt Email Credentials

```bash
# Use environment variables instead of config file
sudo nano /etc/systemd/system/media-ingest.service

[Service]
Environment="EMAIL_PASSWORD=your-secure-password"

# Update application to read from environment
# (requires code modification)
```

## Production Deployment Checklist

- [ ] Service starts automatically on boot
- [ ] Logs are being written correctly
- [ ] Device detection is working
- [ ] File organization is correct
- [ ] Checksums are being verified
- [ ] Email notifications are working (if enabled)
- [ ] Sufficient disk space allocated
- [ ] Backup strategy in place
- [ ] Monitoring is configured
- [ ] Documentation is accessible
- [ ] Team is trained on usage

## Maintenance Schedule

### Daily
- Monitor disk space
- Check for failed transfers in logs

### Weekly
- Review log files for errors
- Verify service is running
- Check destination folder structure

### Monthly
- Clean old logs (if not auto-rotating)
- Review and update configuration
- Check for software updates
- Verify backup strategy

### Quarterly
- Review security settings
- Update dependencies
- Performance audit
- Disaster recovery test

## Support and Debugging

### Enable Debug Logging

```bash
# Edit config
sudo nano /etc/media-ingest/config.yaml

# Change log level
logging:
  log_level: "debug"

# Restart service
sudo systemctl restart media-ingest
```

### Collect Diagnostic Information

```bash
#!/bin/bash
# Save as collect-diagnostics.sh

echo "Collecting diagnostic information..."

{
  echo "=== System Information ==="
  uname -a
  lsb_release -a
  
  echo -e "\n=== Service Status ==="
  systemctl status media-ingest
  
  echo -e "\n=== Recent Logs ==="
  journalctl -u media-ingest -n 100 --no-pager
  
  echo -e "\n=== Configuration ==="
  cat /etc/media-ingest/config.yaml
  
  echo -e "\n=== Disk Usage ==="
  df -h
  
  echo -e "\n=== Block Devices ==="
  lsblk
  
  echo -e "\n=== Mount Points ==="
  mount | grep media
  
  echo -e "\n=== udev Rules ==="
  cat /etc/udev/rules.d/99-media-ingest.rules
  
} > media-ingest-diagnostics-$(date +%Y%m%d-%H%M%S).txt

echo "Diagnostics saved to media-ingest-diagnostics-*.txt"
```

## Contact and Support

For issues not covered in this guide:
1. Check the logs: `sudo journalctl -u media-ingest -f`
2. Review the troubleshooting section
3. Search existing issues on GitHub
4. Open a new issue with diagnostic information
