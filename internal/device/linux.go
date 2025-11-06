// +build linux

package device

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/autofileingest/internal/config"
	"github.com/autofileingest/internal/logger"
)

// LinuxDetector implements DeviceDetector for Linux
type LinuxDetector struct {
	config   *config.Config
	logger   *logger.Logger
	stopChan chan struct{}
	watching bool
}

// NewLinuxDetector creates a new Linux device detector
func NewLinuxDetector(cfg *config.Config, log *logger.Logger) *LinuxDetector {
	return &LinuxDetector{
		config:   cfg,
		logger:   log,
		stopChan: make(chan struct{}),
	}
}

// DetectDevices scans for available block devices
func (l *LinuxDetector) DetectDevices() ([]*Device, error) {
	devices := []*Device{}

	// Use lsblk to list block devices
	cmd := exec.Command("lsblk", "-J", "-o", "NAME,SIZE,TYPE,MOUNTPOINT,FSTYPE,LABEL")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list block devices: %w", err)
	}

	// Parse lsblk output (simplified - in production use proper JSON parsing)
	l.logger.Debug("Detected devices: %s", string(output))

	return devices, nil
}

// MountDevice mounts a device to the configured mount point
func (l *LinuxDetector) MountDevice(device *Device) error {
	if !l.config.AutoMount.Enabled {
		return fmt.Errorf("auto-mount is disabled")
	}

	// Create mount point
	mountPath := filepath.Join(l.config.AutoMount.MountBase, device.Name)
	if err := os.MkdirAll(mountPath, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %w", err)
	}

	// Mount the device
	cmd := exec.Command("mount", device.Path, mountPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mount device: %w", err)
	}

	device.MountPath = mountPath
	l.logger.Success("Mounted device %s at %s", device.Name, mountPath)

	return nil
}

// UnmountDevice unmounts a device
func (l *LinuxDetector) UnmountDevice(device *Device) error {
	if device.MountPath == "" {
		return nil
	}

	cmd := exec.Command("umount", device.MountPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unmount device: %w", err)
	}

	l.logger.Info("Unmounted device %s from %s", device.Name, device.MountPath)
	device.MountPath = ""

	return nil
}

// GetDeviceInfo retrieves detailed information about a block device
func (l *LinuxDetector) GetDeviceInfo(devicePath string) (*Device, error) {
	device := &Device{
		Path: devicePath,
		Name: filepath.Base(devicePath),
	}

	// Get filesystem type
	cmd := exec.Command("blkid", "-s", "TYPE", "-o", "value", devicePath)
	if output, err := cmd.Output(); err == nil {
		device.Filesystem = strings.TrimSpace(string(output))
	}

	// Get label
	cmd = exec.Command("blkid", "-s", "LABEL", "-o", "value", devicePath)
	if output, err := cmd.Output(); err == nil {
		device.Label = strings.TrimSpace(string(output))
	}

	// Get size
	cmd = exec.Command("blockdev", "--getsize64", devicePath)
	if output, err := cmd.Output(); err == nil {
		fmt.Sscanf(string(output), "%d", &device.Size)
	}

	return device, nil
}

// WatchForDevices watches for new devices (simplified for this implementation)
func (l *LinuxDetector) WatchForDevices(callback func(*Device)) error {
	l.watching = true
	l.logger.Info("Started watching for block devices on Linux")
	// In a full implementation, this would use udev or inotify
	return nil
}

// StopWatching stops watching for devices
func (l *LinuxDetector) StopWatching() {
	if l.watching {
		close(l.stopChan)
		l.stopChan = make(chan struct{})
		l.watching = false
	}
}

// GetMountedDevices returns a list of currently mounted devices
func (l *LinuxDetector) GetMountedDevices() ([]*Device, error) {
	devices := []*Device{}

	// Read /proc/mounts
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		
		if len(fields) < 3 {
			continue
		}

		devicePath := fields[0]
		mountPath := fields[1]
		fsType := fields[2]

		// Skip non-device mounts
		if !strings.HasPrefix(devicePath, "/dev/") {
			continue
		}

		device := &Device{
			Path:       devicePath,
			Name:       filepath.Base(devicePath),
			MountPath:  mountPath,
			Filesystem: fsType,
		}

		devices = append(devices, device)
	}

	return devices, nil
}
