# Media Ingest Server

A high-performance, automated media ingest system for Linux that detects SD cards and SSDs, automatically transfers files, and organizes them based on filename patterns.

## Features

✨ **Automatic Device Detection**
- Detects SD cards and SSDs when plugged in using udev rules
- Automatically mounts devices in plug-and-play fashion
- Supports multiple devices connected simultaneously
- Each device processed independently and concurrently

📁 **Intelligent File Organization**
- Parses filenames following pattern: `ProjectName_Client_ACam_ClipNumber.mp4`
- Automatically creates nested folder structure: `[Client]/[ProjectName]/[ACam|BCam|CCam]/`
- Handles files that don't match pattern (moved to "Unsorted" folder)
- Automatic file versioning for duplicates (`filename_v2.mp4`, etc.)

🚀 **High-Performance Transfer**
- Concurrent file transfers using worker pools
- Priority queue system (files starting with `1_` transferred first)
- Real-time progress display with speed, file count, and percentage
- SHA256 checksum verification for file integrity
- Efficient handling of large video files (50GB+)

📊 **Comprehensive Logging**
- Detailed logs stored on server and source device
- Transfer statistics (speed, file count, duration)
- Checksum verification results
- Files that couldn't be parsed
- Errors and warnings

📧 **Email Notifications** (Optional)
- Configurable SMTP settings
- Transfer summary with statistics
- Log file attachments

## Installation

### Prerequisites

- Linux system (Ubuntu, Debian, Fedora, CentOS, Arch, etc.)
- Go 1.21 or higher
- Root/sudo access
- udev support

### Quick Install

```bash
# Clone or download the repository
cd AutoFileIngest

# Run the installation script
sudo chmod +x install.sh
sudo ./install.sh
```

The installation script will:
1. Install dependencies (Go, git, udev)
2. Build the application
3. Install binary to `/usr/local/bin/media-ingest`
4. Create configuration directory at `/etc/media-ingest/`
5. Set up log directory at `/var/log/media-ingest/`
6. Create mount directory at `/mnt/ingest/`
7. Install systemd service
8. Install udev rules

### Manual Installation

If you prefer to install manually:

```bash
# Install dependencies
sudo apt-get update
sudo apt-get install -y golang-go git udev

# Build the application
go mod download
go build -o media-ingest ./cmd/media-ingest

# Install
sudo install -m 755 media-ingest /usr/local/bin/media-ingest
sudo mkdir -p /etc/media-ingest /var/log/media-ingest /mnt/ingest
sudo cp config.example.yaml /etc/media-ingest/config.yaml
sudo cp systemd/media-ingest.service /etc/systemd/system/
sudo cp udev/99-media-ingest.rules /etc/udev/rules.d/

# Reload services
sudo systemctl daemon-reload
sudo udevadm control --reload-rules
sudo udevadm trigger
```

## Configuration

Edit the configuration file at `/etc/media-ingest/config.yaml`:

```yaml
# Destination path where files will be organized
destination_path: "/mnt/storage/media"

# Auto-mount configuration
auto_mount:
  mount_base: "/mnt/ingest"
  enabled: true

# Logging
logging:
  server_log_path: "/var/log/media-ingest"
  log_to_device: true
  retention_days: 90
  log_level: "info"

# File transfer settings
transfer:
  max_workers: 4
  buffer_size: 1048576  # 1MB
  verify_checksums: true
  max_retries: 3
  priority_prefixes:
    - "1_"

# Filename parsing
parsing:
  pattern: "^([^_]+)_([^_]+)_(ACam|BCam|CCam)_(.+)$"
  folder_structure: "{client}/{project}/{camera}"
  unmatched_folder: "Unsorted"

# Email notifications (optional)
email:
  enabled: false
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  use_tls: true
  username: "your-email@gmail.com"
  password: "your-app-password"
  from: "media-ingest@example.com"
  to:
    - "admin@example.com"
```

## Usage

### Running as a Service

```bash
# Start the service
sudo systemctl start media-ingest

# Enable auto-start on boot
sudo systemctl enable media-ingest

# Check status
sudo systemctl status media-ingest

# View live logs
sudo journalctl -u media-ingest -f

# Stop the service
sudo systemctl stop media-ingest
```

### Running Manually

For testing or debugging:

```bash
sudo media-ingest -config /etc/media-ingest/config.yaml
```

### How It Works

1. **Device Detection**: When you plug in an SD card or SSD, udev detects it
2. **Auto-Mount**: The device is automatically mounted to `/mnt/ingest/[device-name]`
3. **File Scanning**: All files on the device are scanned
4. **Prioritization**: Files starting with `1_` are queued first
5. **Transfer**: Files are copied (not moved) to the destination with checksum verification
6. **Organization**: Files are organized based on the filename pattern into nested folders
7. **Logging**: Detailed logs are created on both server and device
8. **Notification**: Optional email notification is sent
9. **Complete**: Device remains mounted for manual verification

### Example File Organization

Given these files on an SD card:
```
1_BrandVideo_Nike_ACam_001.mp4
1_BrandVideo_Nike_BCam_001.mp4
ProductShoot_Adidas_ACam_042.mp4
ProductShoot_Adidas_CCam_043.mp4
random_file.txt
```

They will be organized as:
```
/mnt/storage/media/
├── Nike/
│   └── BrandVideo/
│       ├── ACam/
│       │   └── 001.mp4
│       └── BCam/
│           └── 001.mp4
├── Adidas/
│   └── ProductShoot/
│       ├── ACam/
│       │   └── 042.mp4
│       └── CCam/
│           └── 043.mp4
└── Unsorted/
    └── random_file.txt
```

## Filename Pattern

The default pattern expects filenames in this format:
```
ProjectName_Client_CameraDesignation_ClipNumber.extension
```

**Components:**
- `ProjectName`: Name of the project (e.g., "BrandVideo", "ProductShoot")
- `Client`: Client name (e.g., "Nike", "Adidas")
- `CameraDesignation`: Must be exactly "ACam", "BCam", or "CCam"
- `ClipNumber`: Clip identifier and extension (e.g., "001.mp4")

**Examples:**
- ✅ `CommercialShoot_Tesla_ACam_Scene01.mp4`
- ✅ `1_ProductLaunch_Apple_BCam_Take05.mov`
- ✅ `Interview_Microsoft_CCam_Part3.mxf`
- ❌ `my_video.mp4` (will go to Unsorted folder)

You can customize the pattern in the config file using regex.

## Logs

### Server Logs
Located at `/var/log/media-ingest/`:
- `server_YYYYMMDD_HHMMSS.log`: Main server log

### Device Logs
Created on each ingested device:
- `ingest_log_YYYYMMDD_HHMMSS_[device-name].txt`: Per-device transfer log

### Log Contents
Each log includes:
- Transfer start/end timestamps
- Device information (name, capacity, filesystem)
- Each file transferred with source/destination paths
- Checksum verification results
- Parsing errors ("confused" files)
- Duplicate files with version numbers
- Transfer speeds and total time
- Any errors or warnings

### Viewing Logs

```bash
# System logs (journalctl)
sudo journalctl -u media-ingest -f

# Server logs
sudo tail -f /var/log/media-ingest/server_*.log

# Latest device log
sudo ls -lt /var/log/media-ingest/ | head -n 2
```

## Testing

### Test Scenarios Covered

1. ✅ Multiple devices plugged in simultaneously
2. ✅ Device unplugged during transfer (graceful handling)
3. ✅ Duplicate filenames in same batch (auto-versioning)
4. ✅ Files that don't match naming convention (Unsorted folder)
5. ✅ Very large files (>50GB with progress tracking)
6. ✅ Disk space issues (error handling)
7. ✅ Checksum verification (integrity assurance)
8. ✅ Priority file transfer (1_ prefix first)

### Manual Testing

```bash
# Create test files
mkdir -p /tmp/test-sd-card
echo "test" > /tmp/test-sd-card/1_TestProject_TestClient_ACam_001.mp4
echo "test" > /tmp/test-sd-card/TestProject_TestClient_BCam_002.mp4

# Simulate mount
sudo mkdir -p /mnt/ingest/sdb1
sudo mount --bind /tmp/test-sd-card /mnt/ingest/sdb1

# Watch the logs
sudo journalctl -u media-ingest -f
```

## Troubleshooting

### Service won't start
```bash
# Check service status
sudo systemctl status media-ingest

# Check logs for errors
sudo journalctl -u media-ingest -n 50
```

### Device not detected
```bash
# Check udev rules
sudo udevadm control --reload-rules
sudo udevadm trigger

# Monitor udev events
sudo udevadm monitor

# Check device detection
lsblk
```

### Mounting issues
```bash
# Check mount permissions
ls -la /mnt/ingest/

# Manually mount for testing
sudo mount /dev/sdb1 /mnt/ingest/test-device
```

### Checksum failures
- Verify source media is not corrupted
- Check disk space on destination
- Review logs for specific errors

## Performance Tuning

### Adjust Worker Count
Edit `/etc/media-ingest/config.yaml`:
```yaml
transfer:
  max_workers: 8  # Increase for faster transfers (uses more CPU)
```

### Buffer Size
```yaml
transfer:
  buffer_size: 4194304  # 4MB (larger = faster for big files)
```

### Disable Checksum Verification
```yaml
transfer:
  verify_checksums: false  # Faster but less safe
```

## Security Considerations

- Service runs as root (required for mounting)
- Uses systemd security features (PrivateTmp, NoNewPrivileges)
- Email passwords stored in config file (use app-specific passwords)
- Logs may contain sensitive filenames

## Uninstall

```bash
# Stop and disable service
sudo systemctl stop media-ingest
sudo systemctl disable media-ingest

# Remove files
sudo rm /usr/local/bin/media-ingest
sudo rm /etc/systemd/system/media-ingest.service
sudo rm /etc/udev/rules.d/99-media-ingest.rules
sudo rm -rf /etc/media-ingest
sudo rm -rf /var/log/media-ingest

# Reload services
sudo systemctl daemon-reload
sudo udevadm control --reload-rules
```

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

MIT License - See LICENSE file for details

## Support

For issues, questions, or feature requests, please open an issue on GitHub.

## Changelog

### Version 1.0.0
- Initial release
- Automatic device detection
- Concurrent file transfers
- Checksum verification
- Priority queue system
- Email notifications
- Comprehensive logging
- Systemd service integration
- udev rules for auto-detection

---

**Built with ❤️ for media professionals**
