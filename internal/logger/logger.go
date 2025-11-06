package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/autofileingest/internal/config"
	"github.com/fatih/color"
)

// Logger handles all logging operations
type Logger struct {
	config     *config.Config
	serverLog  *os.File
	deviceLogs map[string]*os.File
	mu         sync.RWMutex
}

// NewLogger creates a new logger instance
func NewLogger(cfg *config.Config) (*Logger, error) {
	// Create server log directory
	if err := os.MkdirAll(cfg.Logging.ServerLogPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create server log file
	logFile := filepath.Join(cfg.Logging.ServerLogPath, fmt.Sprintf("server_%s.log", time.Now().Format("20060102_150405")))
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return &Logger{
		config:     cfg,
		serverLog:  f,
		deviceLogs: make(map[string]*os.File),
	}, nil
}

// Close closes all log files
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.serverLog != nil {
		l.serverLog.Close()
	}

	for _, f := range l.deviceLogs {
		f.Close()
	}

	return nil
}

// CreateDeviceLog creates a log file for a specific device
func (l *Logger) CreateDeviceLog(deviceName, mountPath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("20060102_150405")
	logFileName := fmt.Sprintf("ingest_log_%s_%s.txt", timestamp, deviceName)

	// Log to device if enabled
	if l.config.Logging.LogToDevice && mountPath != "" {
		deviceLogPath := filepath.Join(mountPath, logFileName)
		f, err := os.OpenFile(deviceLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to create device log: %w", err)
		}
		l.deviceLogs[deviceName] = f
	}

	return nil
}

// CloseDeviceLog closes the log file for a specific device
func (l *Logger) CloseDeviceLog(deviceName string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if f, ok := l.deviceLogs[deviceName]; ok {
		f.Close()
		delete(l.deviceLogs, deviceName)
	}
}

// log writes to both server log and device log if available
func (l *Logger) log(level, deviceName, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)

	// Write to server log
	l.mu.RLock()
	if l.serverLog != nil {
		l.serverLog.WriteString(logLine)
	}

	// Write to device log if available
	if deviceName != "" {
		if f, ok := l.deviceLogs[deviceName]; ok {
			f.WriteString(logLine)
		}
	}
	l.mu.RUnlock()
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log("INFO", "", format, args...)
	if l.config.Performance.ColoredOutput {
		color.White("[INFO] "+format, args...)
	} else {
		fmt.Printf("[INFO] "+format+"\n", args...)
	}
}

// Success logs a success message
func (l *Logger) Success(format string, args ...interface{}) {
	l.log("SUCCESS", "", format, args...)
	if l.config.Performance.ColoredOutput {
		color.Green("[SUCCESS] "+format, args...)
	} else {
		fmt.Printf("[SUCCESS] "+format+"\n", args...)
	}
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	l.log("WARNING", "", format, args...)
	if l.config.Performance.ColoredOutput {
		color.Yellow("[WARNING] "+format, args...)
	} else {
		fmt.Printf("[WARNING] "+format+"\n", args...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log("ERROR", "", format, args...)
	if l.config.Performance.ColoredOutput {
		color.Red("[ERROR] "+format, args...)
	} else {
		fmt.Printf("[ERROR] "+format+"\n", args...)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.config.Logging.LogLevel == "debug" {
		l.log("DEBUG", "", format, args...)
		if l.config.Performance.ColoredOutput {
			color.Cyan("[DEBUG] "+format, args...)
		} else {
			fmt.Printf("[DEBUG] "+format+"\n", args...)
		}
	}
}

// DeviceInfo logs device-specific info
func (l *Logger) DeviceInfo(deviceName, format string, args ...interface{}) {
	l.log("INFO", deviceName, format, args...)
	if l.config.Performance.ColoredOutput {
		color.White("[%s] "+format, append([]interface{}{deviceName}, args...)...)
	} else {
		fmt.Printf("[%s] "+format+"\n", append([]interface{}{deviceName}, args...)...)
	}
}

// DeviceError logs device-specific error
func (l *Logger) DeviceError(deviceName, format string, args ...interface{}) {
	l.log("ERROR", deviceName, format, args...)
	if l.config.Performance.ColoredOutput {
		color.Red("[%s] [ERROR] "+format, append([]interface{}{deviceName}, args...)...)
	} else {
		fmt.Printf("[%s] [ERROR] "+format+"\n", append([]interface{}{deviceName}, args...)...)
	}
}

// DeviceSuccess logs device-specific success
func (l *Logger) DeviceSuccess(deviceName, format string, args ...interface{}) {
	l.log("SUCCESS", deviceName, format, args...)
	if l.config.Performance.ColoredOutput {
		color.Green("[%s] [SUCCESS] "+format, append([]interface{}{deviceName}, args...)...)
	} else {
		fmt.Printf("[%s] [SUCCESS] "+format+"\n", append([]interface{}{deviceName}, args...)...)
	}
}

// GetDeviceLogWriter returns a writer for device-specific logs
func (l *Logger) GetDeviceLogWriter(deviceName string) io.Writer {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if f, ok := l.deviceLogs[deviceName]; ok {
		return io.MultiWriter(l.serverLog, f)
	}
	return l.serverLog
}
