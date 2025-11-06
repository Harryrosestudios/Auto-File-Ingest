// +build !windows

package device

import (
	"github.com/autofileingest/internal/config"
	"github.com/autofileingest/internal/logger"
)

// NewWindowsDetector stub for non-Windows platforms
func NewWindowsDetector(cfg *config.Config, log *logger.Logger) DeviceDetector {
	log.Warning("Windows device detection not available on this platform")
	return nil
}
