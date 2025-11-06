package device

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/autofileingest/internal/config"
	"github.com/autofileingest/internal/logger"
	"github.com/autofileingest/internal/parser"
	"github.com/autofileingest/internal/transfer"
)

// DeviceDetector interface for platform-specific device detection
type DeviceDetector interface {
	DetectDevices() ([]*Device, error)
	MountDevice(device *Device) error
	UnmountDevice(device *Device) error
	GetDeviceInfo(devicePath string) (*Device, error)
	WatchForDevices(callback func(*Device)) error
	StopWatching()
}

// Device represents a connected storage device
type Device struct {
	Name       string
	Path       string
	MountPath  string
	Filesystem string
	Size       int64
	Label      string
}

// Manager handles device operations (platform-agnostic)
type Manager struct {
	config         *config.Config
	logger         *logger.Logger
	parser         *parser.Parser
	detector       DeviceDetector
	activeDevices  map[string]*Device
	mu             sync.RWMutex
}

// NewManager creates a new device manager with platform-specific detector
func NewManager(cfg *config.Config, log *logger.Logger) *Manager {
	p, err := parser.NewParser(cfg)
	if err != nil {
		log.Error("Failed to create parser: %v", err)
		return nil
	}

	// Create platform-specific detector
	var detector DeviceDetector
	if runtime.GOOS == "windows" {
		detector = NewWindowsDetector(cfg, log)
	} else { // linux, darwin, etc.
		detector = NewLinuxDetector(cfg, log)
	}

	return &Manager{
		config:        cfg,
		logger:        log,
		parser:        p,
		detector:      detector,
		activeDevices: make(map[string]*Device),
	}
}

// DetectDevices scans for available devices
func (m *Manager) DetectDevices() ([]*Device, error) {
	return m.detector.DetectDevices()
}

// MountDevice mounts a device
func (m *Manager) MountDevice(device *Device) error {
	return m.detector.MountDevice(device)
}

// UnmountDevice unmounts a device
func (m *Manager) UnmountDevice(device *Device) error {
	return m.detector.UnmountDevice(device)
}

// GetDeviceInfo retrieves device information
func (m *Manager) GetDeviceInfo(devicePath string) (*Device, error) {
	return m.detector.GetDeviceInfo(devicePath)
}

// ProcessDevice handles the complete ingest workflow for a device
func (m *Manager) ProcessDevice(device *Device) error {
	m.mu.Lock()
	m.activeDevices[device.Name] = device
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.activeDevices, device.Name)
		m.mu.Unlock()
	}()

	m.logger.Info("Processing device: %s (%s)", device.Name, device.Label)

	// Create device log
	if err := m.logger.CreateDeviceLog(device.Name, device.MountPath); err != nil {
		m.logger.Warning("Failed to create device log: %v", err)
	}
	defer m.logger.CloseDeviceLog(device.Name)

	// Scan for files
	files, err := m.scanFiles(device.MountPath)
	if err != nil {
		m.logger.DeviceError(device.Name, "Failed to scan files: %v", err)
		return err
	}

	m.logger.DeviceInfo(device.Name, "Found %d files to transfer", len(files))

	if len(files) == 0 {
		m.logger.DeviceInfo(device.Name, "No files to transfer")
		return nil
	}

	// Create transfer manager
	transferMgr := transfer.NewManager(m.config, m.logger, m.parser)

	// Start transfer
	if err := transferMgr.TransferFiles(device.Name, files); err != nil {
		m.logger.DeviceError(device.Name, "Transfer failed: %v", err)
		return err
	}

	// Get final statistics
	stats := transferMgr.GetStats()
	m.logger.DeviceSuccess(device.Name, "Transfer complete: %d/%d files transferred",
		stats.ProcessedFiles-stats.FailedFiles, stats.TotalFiles)

	return nil
}

// scanFiles recursively scans for all files in a directory
func (m *Manager) scanFiles(rootPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// IsAllowedDevice checks if a device should be processed
func (m *Manager) IsAllowedDevice(device *Device) bool {
	// Check minimum size
	if device.Size < m.config.DeviceDetection.MinSizeBytes {
		return false
	}

	// Check filesystem
	if len(m.config.DeviceDetection.AllowedFilesystems) > 0 {
		allowed := false
		for _, fs := range m.config.DeviceDetection.AllowedFilesystems {
			if strings.EqualFold(device.Filesystem, fs) || device.Filesystem == fs {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	// Check exclude patterns
	for _, pattern := range m.config.DeviceDetection.ExcludePatterns {
		if strings.Contains(device.Path, pattern) {
			return false
		}
	}

	return true
}

// WatchForDevices starts watching for new devices
func (m *Manager) WatchForDevices(callback func(*Device)) error {
	return m.detector.WatchForDevices(callback)
}

// StopWatching stops watching for devices
func (m *Manager) StopWatching() {
	m.detector.StopWatching()
}

// GetActiveDevices returns currently processing devices
func (m *Manager) GetActiveDevices() []*Device {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]*Device, 0, len(m.activeDevices))
	for _, dev := range m.activeDevices {
		devices = append(devices, dev)
	}
	return devices
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
