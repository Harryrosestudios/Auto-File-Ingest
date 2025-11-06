// +build !linux

package device

import (
	"github.com/autofileingest/internal/config"
	"github.com/autofileingest/internal/logger"
)

// NewLinuxDetector stub for non-Linux platforms
func NewLinuxDetector(cfg *config.Config, log *logger.Logger) DeviceDetector {
	log.Warning("Linux device detection not available on this platform")
	return nil
}
