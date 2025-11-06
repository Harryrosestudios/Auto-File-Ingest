package monitor

import (
	"fmt"
	"runtime"
	"time"

	"github.com/autofileingest/internal/config"
	"github.com/autofileingest/internal/device"
	"github.com/autofileingest/internal/logger"
)

// Monitor watches for device changes
type Monitor struct {
	config     *config.Config
	logger     *logger.Logger
	deviceMgr  *device.Manager
	stopChan   chan struct{}
}

// NewMonitor creates a new device monitor
func NewMonitor(cfg *config.Config, log *logger.Logger, deviceMgr *device.Manager) (*Monitor, error) {
	return &Monitor{
		config:    cfg,
		logger:    log,
		deviceMgr: deviceMgr,
		stopChan:  make(chan struct{}),
	}, nil
}

// Start begins monitoring for device changes
func (m *Monitor) Start() error {
	m.logger.Info("Device monitoring started on %s", runtime.GOOS)

	// Start device watching
	err := m.deviceMgr.WatchForDevices(func(dev *device.Device) {
		m.handleDeviceAdded(dev)
	})
	
	if err != nil {
		return fmt.Errorf("failed to start device watching: %w", err)
	}

	// Initial scan for already connected devices
	go m.scanExistingDevices()

	return nil
}

// Stop stops the monitor
func (m *Monitor) Stop() {
	close(m.stopChan)
	m.deviceMgr.StopWatching()
	m.logger.Info("Device monitoring stopped")
}

// handleDeviceAdded processes newly added devices
func (m *Monitor) handleDeviceAdded(dev *device.Device) {
	m.logger.Debug("New device detected: %s", dev.Path)

	// Wait a moment for device to be ready
	time.Sleep(2 * time.Second)

	// Check if device should be processed
	if !m.deviceMgr.IsAllowedDevice(dev) {
		m.logger.Debug("Device %s not allowed (size: %d, fs: %s)", dev.Name, dev.Size, dev.Filesystem)
		return
	}

	m.logger.Info("New device detected: %s (%s, %s)", dev.Name, dev.Label, formatSize(dev.Size))

	// Mount device
	if err := m.deviceMgr.MountDevice(dev); err != nil {
		m.logger.Error("Failed to mount device %s: %v", dev.Name, err)
		return
	}

	// Process device in background
	go func() {
		if err := m.deviceMgr.ProcessDevice(dev); err != nil {
			m.logger.Error("Failed to process device %s: %v", dev.Name, err)
		}
		// Keep device mounted after processing (as per requirements)
	}()
}

// scanExistingDevices scans for devices that are already connected
func (m *Monitor) scanExistingDevices() {
	m.logger.Info("Scanning for existing devices...")

	devices, err := m.deviceMgr.DetectDevices()
	if err != nil {
		m.logger.Error("Failed to detect devices: %v", err)
		return
	}

	for _, dev := range devices {
		if m.deviceMgr.IsAllowedDevice(dev) {
			m.logger.Info("Found existing device: %s (%s)", dev.Name, dev.Label)
			
			// Mount if needed
			if dev.MountPath == "" {
				if err := m.deviceMgr.MountDevice(dev); err != nil {
					m.logger.Error("Failed to mount device %s: %v", dev.Name, err)
					continue
				}
			}
			
			// Process device
			go func(d *device.Device) {
				if err := m.deviceMgr.ProcessDevice(d); err != nil {
					m.logger.Error("Failed to process device %s: %v", d.Name, err)
				}
			}(dev)
		}
	}
}

// formatSize formats bytes as human-readable size
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
