// +build windows

package device

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/autofileingest/internal/config"
	"github.com/autofileingest/internal/logger"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	getDriveTypeW      = kernel32.NewProc("GetDriveTypeW")
	getVolumeInformationW = kernel32.NewProc("GetVolumeInformationW")
	getDiskFreeSpaceExW = kernel32.NewProc("GetDiskFreeSpaceExW")
)

const (
	DRIVE_REMOVABLE = 2
	DRIVE_FIXED     = 3
)

// WindowsDetector implements DeviceDetector for Windows
type WindowsDetector struct {
	config    *config.Config
	logger    *logger.Logger
	stopChan  chan struct{}
	watching  bool
	knownDrives map[string]bool
}

// NewWindowsDetector creates a new Windows device detector
func NewWindowsDetector(cfg *config.Config, log *logger.Logger) *WindowsDetector {
	return &WindowsDetector{
		config:      cfg,
		logger:      log,
		stopChan:    make(chan struct{}),
		knownDrives: make(map[string]bool),
	}
}

// DetectDevices scans for removable drives on Windows
func (w *WindowsDetector) DetectDevices() ([]*Device, error) {
	devices := []*Device{}

	// Get all drive letters
	drives := w.getLogicalDrives()

	for _, drive := range drives {
		driveType := w.getDriveType(drive)
		
		// Only process removable drives (SD cards, USB)
		if driveType == DRIVE_REMOVABLE {
			device, err := w.GetDeviceInfo(drive)
			if err != nil {
				w.logger.Debug("Failed to get info for drive %s: %v", drive, err)
				continue
			}
			devices = append(devices, device)
		}
	}

	return devices, nil
}

// MountDevice is a no-op on Windows (drives are auto-mounted)
func (w *WindowsDetector) MountDevice(device *Device) error {
	// On Windows, removable drives are automatically mounted
	// We just verify the drive exists and is accessible
	if _, err := os.Stat(device.Path); err != nil {
		return fmt.Errorf("drive not accessible: %w", err)
	}
	
	// Set mount path to the drive letter itself
	device.MountPath = device.Path
	w.logger.Success("Drive %s ready at %s", device.Name, device.MountPath)
	
	return nil
}

// UnmountDevice is a no-op on Windows
func (w *WindowsDetector) UnmountDevice(device *Device) error {
	// Windows handles unmounting through the system
	w.logger.Info("Drive %s can be safely removed", device.Name)
	return nil
}

// GetDeviceInfo retrieves information about a Windows drive
func (w *WindowsDetector) GetDeviceInfo(drivePath string) (*Device, error) {
	// Ensure drive path ends with backslash
	if !strings.HasSuffix(drivePath, "\\") {
		drivePath += "\\"
	}

	device := &Device{
		Name: filepath.VolumeName(drivePath),
		Path: drivePath,
	}

	// Get volume information
	var volumeNameBuffer [syscall.MAX_PATH + 1]uint16
	var fileSystemNameBuffer [syscall.MAX_PATH + 1]uint16
	var volumeSerialNumber uint32
	var maximumComponentLength uint32
	var fileSystemFlags uint32

	drivePtr, _ := syscall.UTF16PtrFromString(drivePath)
	ret, _, _ := getVolumeInformationW.Call(
		uintptr(unsafe.Pointer(drivePtr)),
		uintptr(unsafe.Pointer(&volumeNameBuffer[0])),
		uintptr(len(volumeNameBuffer)),
		uintptr(unsafe.Pointer(&volumeSerialNumber)),
		uintptr(unsafe.Pointer(&maximumComponentLength)),
		uintptr(unsafe.Pointer(&fileSystemFlags)),
		uintptr(unsafe.Pointer(&fileSystemNameBuffer[0])),
		uintptr(len(fileSystemNameBuffer)),
	)

	if ret != 0 {
		device.Label = syscall.UTF16ToString(volumeNameBuffer[:])
		device.Filesystem = syscall.UTF16ToString(fileSystemNameBuffer[:])
	}

	// Get disk size
	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes int64
	ret, _, _ = getDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(drivePtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)

	if ret != 0 {
		device.Size = totalNumberOfBytes
	}

	return device, nil
}

// WatchForDevices watches for new removable drives
func (w *WindowsDetector) WatchForDevices(callback func(*Device)) error {
	if w.watching {
		return fmt.Errorf("already watching for devices")
	}

	w.watching = true
	w.logger.Info("Started watching for removable drives on Windows")

	// Initialize known drives
	currentDrives := w.getLogicalDrives()
	for _, drive := range currentDrives {
		if w.getDriveType(drive) == DRIVE_REMOVABLE {
			w.knownDrives[drive] = true
		}
	}

	// Poll for new drives
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-w.stopChan:
				w.watching = false
				return
			case <-ticker.C:
				w.checkForNewDrives(callback)
			}
		}
	}()

	return nil
}

// StopWatching stops watching for new devices
func (w *WindowsDetector) StopWatching() {
	if w.watching {
		close(w.stopChan)
		w.stopChan = make(chan struct{})
	}
}

// getLogicalDrives returns all available drive letters
func (w *WindowsDetector) getLogicalDrives() []string {
	drives := []string{}
	
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		drivePath := string(drive) + ":\\"
		if _, err := os.Stat(drivePath); err == nil {
			drives = append(drives, drivePath)
		}
	}
	
	return drives
}

// getDriveType returns the Windows drive type
func (w *WindowsDetector) getDriveType(drivePath string) uint32 {
	drivePtr, _ := syscall.UTF16PtrFromString(drivePath)
	ret, _, _ := getDriveTypeW.Call(uintptr(unsafe.Pointer(drivePtr)))
	return uint32(ret)
}

// checkForNewDrives checks for newly connected drives
func (w *WindowsDetector) checkForNewDrives(callback func(*Device)) {
	currentDrives := w.getLogicalDrives()
	
	for _, drive := range currentDrives {
		driveType := w.getDriveType(drive)
		
		// Only process removable drives
		if driveType == DRIVE_REMOVABLE {
			// Check if this is a new drive
			if !w.knownDrives[drive] {
				w.knownDrives[drive] = true
				w.logger.Info("New removable drive detected: %s", drive)
				
				device, err := w.GetDeviceInfo(drive)
				if err != nil {
					w.logger.Error("Failed to get device info for %s: %v", drive, err)
					continue
				}
				
				// Trigger callback
				callback(device)
			}
		} else {
			// Drive was removed or changed type
			delete(w.knownDrives, drive)
		}
	}
	
	// Remove drives that are no longer present
	for drive := range w.knownDrives {
		found := false
		for _, currentDrive := range currentDrives {
			if drive == currentDrive && w.getDriveType(currentDrive) == DRIVE_REMOVABLE {
				found = true
				break
			}
		}
		if !found {
			w.logger.Info("Drive removed: %s", drive)
			delete(w.knownDrives, drive)
		}
	}
}
