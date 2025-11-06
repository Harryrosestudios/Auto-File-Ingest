# Example Usage Scenarios

This document provides real-world usage examples for Media Ingest Server.

## Scenario 1: Video Production Company

### Setup
A video production company shoots with 3 cameras (A, B, C) for multiple clients.

### Filename Convention
```
ProjectName_ClientName_CameraLetter_ClipNumber.extension
```

### Examples
```
CommercialShoot_Nike_ACam_Scene01.mp4
CommercialShoot_Nike_BCam_Scene01.mp4
CommercialShoot_Nike_CCam_Scene01.mp4
1_Priority_Nike_ACam_Scene02.mp4  # Priority file
ProductLaunch_Apple_ACam_Take01.mov
```

### Expected Output Structure
```
/mnt/storage/media/
├── Nike/
│   └── CommercialShoot/
│       ├── ACam/
│       │   ├── Scene01.mp4
│       │   └── Scene02.mp4
│       ├── BCam/
│       │   └── Scene01.mp4
│       └── CCam/
│           └── Scene01.mp4
└── Apple/
    └── ProductLaunch/
        └── ACam/
            └── Take01.mov
```

### Configuration
```yaml
destination_path: "/mnt/storage/media"

transfer:
  max_workers: 6  # Fast multi-camera transfer
  verify_checksums: true  # Critical for video files
  priority_prefixes:
    - "1_"  # Rush deliveries first

parsing:
  pattern: "^([^_]+)_([^_]+)_(ACam|BCam|CCam)_(.+)$"
  folder_structure: "{client}/{project}/{camera}"

email:
  enabled: true
  to:
    - "editor@company.com"
    - "producer@company.com"
```

## Scenario 2: News Studio

### Setup
News studio with multiple field reporters, each with their own camera.

### Filename Convention
```
StoryName_Reporter_ACam_Date.extension
```

### Examples
```
BreakingNews_JohnDoe_ACam_20250106.mp4
1_LiveEvent_JaneSmith_BCam_20250106.mxf
Interview_MikeJones_ACam_20250106.mov
```

### Configuration
```yaml
destination_path: "/mnt/archive/news"

transfer:
  max_workers: 4
  priority_prefixes:
    - "1_"  # Breaking news first
    - "Breaking"  # Additional priority

parsing:
  pattern: "^([^_]+)_([^_]+)_(ACam|BCam|CCam)_(.+)$"
  folder_structure: "{client}/{project}/{camera}"

email:
  enabled: true
  subject: "News Footage Available - {device}"
  to:
    - "newsroom@studio.com"
```

## Scenario 3: Wedding Videography

### Setup
Wedding videographer with 2 cameras covering ceremonies.

### Filename Convention
```
EventName_CoupleNames_Camera_Clip.extension
```

### Examples
```
WeddingCeremony_Smith_ACam_Vows.mp4
WeddingCeremony_Smith_BCam_Vows.mp4
Reception_Smith_ACam_Dance01.mp4
1_HighlightReel_Smith_ACam_Best.mp4  # Priority edit
```

### Configuration
```yaml
destination_path: "/mnt/weddings"

transfer:
  max_workers: 2  # Laptop setup
  buffer_size: 2097152  # 2MB buffer

parsing:
  pattern: "^([^_]+)_([^_]+)_(ACam|BCam|CCam)_(.+)$"
  folder_structure: "{client}/{project}/{camera}"
  unmatched_folder: "ToReview"  # Manual review needed

email:
  enabled: true
  to:
    - "editor@weddings.com"
  subject: "Wedding Footage Ready - {device}"
```

## Scenario 4: Multi-Location Corporate

### Setup
Corporate video team shooting at different locations.

### Filename Convention
```
ProjectName_Location_Camera_Scene.extension
```

### Examples
```
AnnualReport_NYC_ACam_CEO.mp4
AnnualReport_NYC_BCam_CEO.mp4
AnnualReport_LA_ACam_Factory.mov
Training_Chicago_CCam_Demo01.mp4
```

### Configuration
```yaml
destination_path: "/mnt/corporate/video"

transfer:
  max_workers: 8  # Powerful server
  buffer_size: 4194304  # 4MB for large files
  verify_checksums: true

parsing:
  pattern: "^([^_]+)_([^_]+)_(ACam|BCam|CCam)_(.+)$"
  folder_structure: "{client}/{project}/{camera}"

device_detection:
  min_size_bytes: 10737418240  # 10GB minimum (filter small drives)

email:
  enabled: true
  to:
    - "postproduction@corp.com"
    - "itadmin@corp.com"
```

## Scenario 5: High-Volume Event Coverage

### Setup
Multi-day conference with 10+ cameras running simultaneously.

### Filename Convention
```
EventName_Day_Camera_Session.extension
```

### Examples
```
Conference2025_Day1_ACam_Keynote.mp4
Conference2025_Day1_BCam_Keynote.mp4
1_Conference2025_Day1_ACam_MainStage.mp4
Conference2025_Day2_CCam_Workshop.mov
```

### Configuration
```yaml
destination_path: "/mnt/raid/events"

transfer:
  max_workers: 12  # High concurrency
  buffer_size: 8388608  # 8MB buffer
  verify_checksums: true
  priority_prefixes:
    - "1_"
    - "Main"
    - "Keynote"

logging:
  log_level: "info"  # Reduce log verbosity
  retention_days: 30

performance:
  show_progress: true
  progress_interval: 2  # Update every 2 seconds

email:
  enabled: true
  to:
    - "production@events.com"
    - "director@events.com"
    - "backup@events.com"
```

## Scenario 6: Backup Workflow

### Setup
Simple backup of SD cards without organization.

### Configuration
```yaml
destination_path: "/mnt/backup/raw"

auto_mount:
  enabled: true
  mount_base: "/mnt/cards"

transfer:
  max_workers: 2
  verify_checksums: true

parsing:
  pattern: "^.*$"  # Match everything
  folder_structure: "backup_{date}"  # Custom structure
  unmatched_folder: "all_files"

logging:
  log_to_device: true  # Important for backup verification

email:
  enabled: true
  subject: "Backup Complete - {device}"
  attach_log: true
```

## Scenario 7: Mixed Format Production

### Setup
Production using multiple file formats and naming schemes.

### Multiple Patterns Support
```yaml
parsing:
  # Primary pattern
  pattern: "^([^_]+)_([^_]+)_(ACam|BCam|CCam)_(.+)$"
  
  # For files that don't match, use alternative handling
  unmatched_folder: "ToOrganize"

# Can extend with custom logic in code
```

## Log Examples

### Successful Transfer Log
```
[2025-01-06 10:30:15] [INFO] Media Ingest Server v1.0.0 starting...
[2025-01-06 10:30:15] [INFO] Configuration loaded from: /etc/media-ingest/config.yaml
[2025-01-06 10:30:15] [INFO] Media Ingest Server is running.
[2025-01-06 10:35:22] [INFO] New device detected: sdb1 (CANON_SD, 64.0 GB)
[2025-01-06 10:35:24] [SUCCESS] Mounted device sdb1 at /mnt/ingest/sdb1
[2025-01-06 10:35:25] [sdb1] Found 150 files to transfer
[2025-01-06 10:35:25] [sdb1] Found 150 files (25 priority, 125 normal)
[2025-01-06 10:35:26] [sdb1] [SUCCESS] Transferred: 001.mp4 -> Nike/BrandVideo/ACam
[2025-01-06 10:35:28] [sdb1] [SUCCESS] Transferred: 002.mp4 -> Nike/BrandVideo/ACam
[2025-01-06 10:45:30] [sdb1] [SUCCESS] Transfer complete: 150/150 files transferred
```

### Error Log Example
```
[2025-01-06 10:35:30] [sdb1] [ERROR] Failed to open source file random.tmp: permission denied
[2025-01-06 10:36:15] [sdb1] [ERROR] Checksum mismatch for Scene05.mp4
[2025-01-06 10:36:16] [sdb1] [WARNING] Retrying transfer for Scene05.mp4 (attempt 2/3)
[2025-01-06 10:36:45] [sdb1] [SUCCESS] Transferred: Scene05.mp4 -> Nike/BrandVideo/ACam
```

## Performance Benchmarks

### Typical Transfer Speeds

| File Size | Workers | Buffer | Speed | Notes |
|-----------|---------|--------|-------|-------|
| 1GB video | 4 | 1MB | 80 MB/s | USB 3.0 SSD |
| 50GB project | 8 | 4MB | 120 MB/s | NVMe to RAID |
| 100x 100MB | 6 | 1MB | 60 MB/s | Multiple small files |

### Optimization Tips

**For Large Files (>10GB):**
```yaml
transfer:
  max_workers: 2  # Fewer, larger streams
  buffer_size: 8388608  # 8MB buffer
```

**For Many Small Files:**
```yaml
transfer:
  max_workers: 8  # More parallel operations
  buffer_size: 524288  # 512KB buffer
```

## Troubleshooting Examples

### Problem: Files going to Unsorted

**Symptom:** All files end up in Unsorted folder

**Solution:** Check filename pattern

```bash
# Test your pattern
cd AutoFileIngest
go run cmd/media-ingest/main.go -test-pattern "YourFileName.mp4"

# Or check logs for parsing errors
sudo journalctl -u media-ingest | grep "unmatched"
```

### Problem: Slow Transfers

**Symptom:** Transfer speed < 20 MB/s

**Checks:**
1. USB port speed: `lsusb -t`
2. Disk speed: `sudo hdparm -t /dev/sdb`
3. Worker count: Check config
4. CPU usage: `top`

**Solutions:**
- Increase buffer size
- Disable checksum verification temporarily
- Check for disk errors

### Problem: Device Not Detected

**Symptom:** Plugging in SD card does nothing

**Checks:**
```bash
# Check udev rules
cat /etc/udev/rules.d/99-media-ingest.rules

# Monitor udev events
sudo udevadm monitor

# Check device visibility
lsblk

# Check service logs
sudo journalctl -u media-ingest -f
```

## Integration Examples

### With Cloud Upload

After transfer completes, upload to cloud:

```bash
# Create post-transfer script
cat > /usr/local/bin/media-ingest-cloud-upload.sh << 'EOF'
#!/bin/bash
# Upload to cloud after ingest
rclone copy /mnt/storage/media/ remote:backup/media/
EOF

chmod +x /usr/local/bin/media-ingest-cloud-upload.sh
```

### With Notification System

Send Slack notification:

```bash
# Install webhook
curl -X POST -H 'Content-type: application/json' \
  --data '{"text":"Media ingest complete!"}' \
  https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

### With Transcoding Pipeline

Trigger transcoding after ingest:

```bash
# Watch destination folder
inotifywait -m /mnt/storage/media/ -e create |
  while read path action file; do
    ffmpeg -i "$path$file" -c:v h264 "$path${file%.*}.h264.mp4"
  done
```

---

These examples should cover most common use cases. Adapt them to your specific needs!
