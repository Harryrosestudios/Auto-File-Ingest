package transfer

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/autofileingest/internal/config"
	"github.com/autofileingest/internal/logger"
	"github.com/autofileingest/internal/parser"
)

// FileTransfer represents a file to be transferred
type FileTransfer struct {
	SourcePath      string
	DestinationPath string
	FileInfo        *parser.FileInfo
	Size            int64
	Priority        bool
	Checksum        string
}

// TransferStats holds transfer statistics
type TransferStats struct {
	TotalFiles      int
	ProcessedFiles  int
	TotalBytes      int64
	TransferredBytes int64
	FailedFiles     int
	SkippedFiles    int
	StartTime       time.Time
	mu              sync.RWMutex
}

// Manager handles file transfers
type Manager struct {
	config  *config.Config
	logger  *logger.Logger
	parser  *parser.Parser
	stats   *TransferStats
}

// NewManager creates a new transfer manager
func NewManager(cfg *config.Config, log *logger.Logger, p *parser.Parser) *Manager {
	return &Manager{
		config: cfg,
		logger: log,
		parser: p,
		stats: &TransferStats{
			StartTime: time.Now(),
		},
	}
}

// TransferFiles transfers files from source to destination
func (m *Manager) TransferFiles(deviceName string, files []string) error {
	m.stats = &TransferStats{
		StartTime: time.Now(),
	}

	// Parse and categorize files
	priorityFiles := []FileTransfer{}
	normalFiles := []FileTransfer{}

	for _, filePath := range files {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			m.logger.DeviceError(deviceName, "Failed to stat file %s: %v", filePath, err)
			continue
		}

		if fileInfo.IsDir() {
			continue
		}

		parsedInfo := m.parser.Parse(filePath)
		destPath, err := m.parser.GetUniqueDestinationPath(parsedInfo)
		if err != nil {
			m.logger.DeviceError(deviceName, "Failed to get destination path for %s: %v", filePath, err)
			continue
		}

		transfer := FileTransfer{
			SourcePath:      filePath,
			DestinationPath: destPath,
			FileInfo:        parsedInfo,
			Size:            fileInfo.Size(),
			Priority:        m.isPriorityFile(filepath.Base(filePath)),
		}

		m.stats.TotalFiles++
		m.stats.TotalBytes += transfer.Size

		if transfer.Priority {
			priorityFiles = append(priorityFiles, transfer)
		} else {
			normalFiles = append(normalFiles, transfer)
		}
	}

	m.logger.DeviceInfo(deviceName, "Found %d files (%d priority, %d normal)", 
		m.stats.TotalFiles, len(priorityFiles), len(normalFiles))

	// Create worker pool
	jobs := make(chan FileTransfer, m.stats.TotalFiles)
	results := make(chan error, m.stats.TotalFiles)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < m.config.Transfer.MaxWorkers; i++ {
		wg.Add(1)
		go m.worker(deviceName, jobs, results, &wg)
	}

	// Send priority files first
	for _, transfer := range priorityFiles {
		jobs <- transfer
	}

	// Then send normal files
	for _, transfer := range normalFiles {
		jobs <- transfer
	}

	close(jobs)

	// Wait for all workers to finish
	wg.Wait()
	close(results)

	// Collect results
	for err := range results {
		if err != nil {
			m.stats.FailedFiles++
		}
	}

	return nil
}

// worker processes file transfers
func (m *Manager) worker(deviceName string, jobs <-chan FileTransfer, results chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for transfer := range jobs {
		err := m.transferFile(deviceName, transfer)
		results <- err
		
		m.stats.mu.Lock()
		m.stats.ProcessedFiles++
		if err == nil {
			m.stats.TransferredBytes += transfer.Size
		}
		m.stats.mu.Unlock()
	}
}

// transferFile transfers a single file
func (m *Manager) transferFile(deviceName string, transfer FileTransfer) error {
	// Create destination directory
	destDir := filepath.Dir(transfer.DestinationPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		m.logger.DeviceError(deviceName, "Failed to create directory %s: %v", destDir, err)
		return err
	}

	// Open source file
	srcFile, err := os.Open(transfer.SourcePath)
	if err != nil {
		m.logger.DeviceError(deviceName, "Failed to open source file %s: %v", transfer.SourcePath, err)
		return err
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(transfer.DestinationPath)
	if err != nil {
		m.logger.DeviceError(deviceName, "Failed to create destination file %s: %v", transfer.DestinationPath, err)
		return err
	}
	defer destFile.Close()

	// Calculate checksum while copying
	var srcChecksum, destChecksum string
	
	if m.config.Transfer.VerifyChecksums {
		srcHash := sha256.New()
		_, err = io.Copy(io.MultiWriter(destFile, srcHash), srcFile)
		if err != nil {
			m.logger.DeviceError(deviceName, "Failed to copy file %s: %v", transfer.SourcePath, err)
			return err
		}
		srcChecksum = fmt.Sprintf("%x", srcHash.Sum(nil))

		// Verify destination file
		destFile.Seek(0, 0)
		destHash := sha256.New()
		if _, err := io.Copy(destHash, destFile); err != nil {
			m.logger.DeviceError(deviceName, "Failed to verify file %s: %v", transfer.DestinationPath, err)
			return err
		}
		destChecksum = fmt.Sprintf("%x", destHash.Sum(nil))

		if srcChecksum != destChecksum {
			m.logger.DeviceError(deviceName, "Checksum mismatch for %s", transfer.SourcePath)
			os.Remove(transfer.DestinationPath)
			return fmt.Errorf("checksum mismatch")
		}
	} else {
		// Simple copy without verification
		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			m.logger.DeviceError(deviceName, "Failed to copy file %s: %v", transfer.SourcePath, err)
			return err
		}
	}

	// Log successful transfer
	if !transfer.FileInfo.Matched {
		m.logger.DeviceInfo(deviceName, "Transferred (unmatched): %s -> %s", 
			filepath.Base(transfer.SourcePath), transfer.DestinationPath)
	} else {
		m.logger.DeviceSuccess(deviceName, "Transferred: %s -> %s/%s/%s", 
			filepath.Base(transfer.SourcePath), 
			transfer.FileInfo.Client, 
			transfer.FileInfo.ProjectName, 
			transfer.FileInfo.Camera)
	}

	return nil
}

// isPriorityFile checks if a file should be transferred with priority
func (m *Manager) isPriorityFile(fileName string) bool {
	for _, prefix := range m.config.Transfer.PriorityPrefixes {
		if len(fileName) >= len(prefix) && fileName[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// GetStats returns current transfer statistics
func (m *Manager) GetStats() TransferStats {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()
	return *m.stats
}

// GetProgress returns transfer progress as percentage
func (m *Manager) GetProgress() float64 {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()
	
	if m.stats.TotalBytes == 0 {
		return 0
	}
	
	return float64(m.stats.TransferredBytes) / float64(m.stats.TotalBytes) * 100
}

// GetSpeed returns current transfer speed in bytes per second
func (m *Manager) GetSpeed() float64 {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()
	
	elapsed := time.Since(m.stats.StartTime).Seconds()
	if elapsed == 0 {
		return 0
	}
	
	return float64(m.stats.TransferredBytes) / elapsed
}
