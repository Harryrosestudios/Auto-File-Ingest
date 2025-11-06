# Quick Start Guide

Get Media Ingest Server up and running in 5 minutes!

## Prerequisites

- Linux system with sudo access
- Go 1.21+ installed
- At least 2GB free disk space

## Installation (3 steps)

### Step 1: Download

```bash
cd ~
git clone <repository-url> AutoFileIngest
cd AutoFileIngest
```

Or download and extract the zip file.

### Step 2: Install

```bash
chmod +x install.sh
sudo ./install.sh
```

This will:
- Install dependencies
- Build the application
- Set up directories
- Install systemd service
- Configure udev rules

### Step 3: Configure

```bash
sudo nano /etc/media-ingest/config.yaml
```

**Minimum required changes:**

```yaml
# Change this to your desired storage location
destination_path: "/mnt/storage/media"
```

Save and exit (Ctrl+X, Y, Enter)

## Start the Service

```bash
# Start now
sudo systemctl start media-ingest

# Enable auto-start on boot
sudo systemctl enable media-ingest

# Check it's running
sudo systemctl status media-ingest
```

## Test It Out

1. **Plug in an SD card or USB drive**
2. **Watch the logs:**
   ```bash
   sudo journalctl -u media-ingest -f
   ```
3. **You should see:**
   - Device detected
   - Mounting
   - File scanning
   - Transfer progress
   - Completion message

## View Results

Check your destination folder:

```bash
ls -la /mnt/storage/media/
```

You should see organized folders based on your filenames!

## What's Next?

### Configure Email Notifications (Optional)

Edit `/etc/media-ingest/config.yaml`:

```yaml
email:
  enabled: true
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "your-email@gmail.com"
  password: "your-app-password"
  to:
    - "admin@example.com"
```

Restart: `sudo systemctl restart media-ingest`

### Adjust Performance

For faster transfers:

```yaml
transfer:
  max_workers: 8  # Increase based on your CPU cores
  buffer_size: 4194304  # 4MB buffer
```

### Customize File Pattern

The default expects: `ProjectName_Client_ACam_001.mp4`

To change, edit the pattern in config:

```yaml
parsing:
  pattern: "^([^_]+)_([^_]+)_(ACam|BCam|CCam)_(.+)$"
  folder_structure: "{client}/{project}/{camera}"
```

## Common Commands

```bash
# Start/stop service
sudo systemctl start media-ingest
sudo systemctl stop media-ingest
sudo systemctl restart media-ingest

# View logs
sudo journalctl -u media-ingest -f

# Check status
sudo systemctl status media-ingest

# View server logs
sudo tail -f /var/log/media-ingest/server_*.log
```

## Troubleshooting

### Service won't start

```bash
sudo journalctl -u media-ingest -n 50
```

Look for error messages.

### Device not detected

```bash
# Reload udev rules
sudo udevadm control --reload-rules
sudo udevadm trigger

# Check if device is visible
lsblk
```

### Files not transferring

Check logs for errors:
```bash
sudo journalctl -u media-ingest -f
```

Common issues:
- Wrong file pattern
- Insufficient permissions
- Disk full

## Getting Help

- Check README.md for detailed documentation
- Check DEPLOYMENT.md for advanced configuration
- View logs: `sudo journalctl -u media-ingest -f`
- Check GitHub issues

## Uninstall

```bash
sudo systemctl stop media-ingest
sudo systemctl disable media-ingest
sudo rm /usr/local/bin/media-ingest
sudo rm -rf /etc/media-ingest
sudo rm /etc/systemd/system/media-ingest.service
sudo rm /etc/udev/rules.d/99-media-ingest.rules
sudo systemctl daemon-reload
```

---

**That's it! You're ready to start automatically ingesting media files!** 🎉
