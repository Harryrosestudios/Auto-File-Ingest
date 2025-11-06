#!/bin/bash
# This script is called by udev when a new device is detected
# Place in /usr/local/bin/media-ingest-trigger

DEVICE=$1

if [ -z "$DEVICE" ]; then
    echo "Usage: $0 <device>"
    exit 1
fi

# Log the event
logger -t media-ingest "Device detected: $DEVICE"

# The main service will detect the device through fsnotify
# This is just for logging/debugging purposes

# Optionally, you can signal the service to check for new devices
# systemctl restart media-ingest

exit 0
