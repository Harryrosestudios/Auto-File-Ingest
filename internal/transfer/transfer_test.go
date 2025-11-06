package transfer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/autofileingest/internal/config"
	"github.com/autofileingest/internal/logger"
	"github.com/autofileingest/internal/parser"
)

func TestTransferManager_Integration(t *testing.T) {
	// Create temporary directories
	sourceDir, err := ioutil.TempDir("", "media-ingest-source-*")
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	destDir, err := ioutil.TempDir("", "media-ingest-dest-*")
	if err != nil {
		t.Fatalf("Failed to create dest dir: %v", err)
	}
	defer os.RemoveAll(destDir)

	logDir, err := ioutil.TempDir("", "media-ingest-logs-*")
	if err != nil {
		t.Fatalf("Failed to create log dir: %v", err)
	}
	defer os.RemoveAll(logDir)

	// Create test files
	testFiles := []string{
		"Priority_Client1_ACam_001.mp4",  // Priority by name
		"Project1_Client1_ACam_002.mp4",
		"Project1_Client1_BCam_001.mp4",
		"Project2_Client2_CCam_010.mov",
		"unmatched_file.mp4",
	}

	for _, filename := range testFiles {
		filepath := filepath.Join(sourceDir, filename)
		content := []byte("test content for " + filename)
		if err := ioutil.WriteFile(filepath, content, 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create config
	cfg := &config.Config{
		DestinationPath: destDir,
		Logging: config.LoggingConfig{
			ServerLogPath: logDir,
			LogToDevice:   false,
			LogLevel:      "debug",
		},
		Transfer: config.TransferConfig{
			MaxWorkers:       2,
			BufferSize:       1024,
			VerifyChecksums:  true,
			MaxRetries:       3,
			PriorityPrefixes: []string{"Priority_"},
		},
		Parsing: config.ParsingConfig{
			Pattern:         "^([^_]+)_([^_]+)_(ACam|BCam|CCam)_(.+)$",
			FolderStructure: "{client}/{project}/{camera}",
			UnmatchedFolder: "Unsorted",
		},
		Performance: config.PerfConfig{
			ShowProgress:     false,
			ProgressInterval: 1,
			ColoredOutput:    false,
		},
	}

	// Create logger
	log, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	// Create parser
	p, err := parser.NewParser(cfg)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Create transfer manager
	mgr := NewManager(cfg, log, p)

	// Collect all test files
	var files []string
	for _, filename := range testFiles {
		files = append(files, filepath.Join(sourceDir, filename))
	}

	// Execute transfer
	err = mgr.TransferFiles("test-device", files)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	// Verify results
	stats := mgr.GetStats()

	if stats.TotalFiles != len(testFiles) {
		t.Errorf("Expected %d total files, got %d", len(testFiles), stats.TotalFiles)
	}

	if stats.ProcessedFiles != len(testFiles) {
		t.Errorf("Expected %d processed files, got %d", len(testFiles), stats.ProcessedFiles)
	}

	if stats.FailedFiles != 0 {
		t.Errorf("Expected 0 failed files, got %d", stats.FailedFiles)
	}

	// Verify file organization
	expectedStructure := map[string][]string{
		"Client1/Priority/ACam":   {"001.mp4"},
		"Client1/Project1/ACam":   {"002.mp4"},
		"Client1/Project1/BCam":   {"001.mp4"},
		"Client2/Project2/CCam":   {"010.mov"},
		"Unsorted":                {"unmatched_file.mp4"},
	}

	for dir, expectedFiles := range expectedStructure {
		fullDir := filepath.Join(destDir, dir)
		if _, err := os.Stat(fullDir); os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", dir)
			continue
		}

		files, err := ioutil.ReadDir(fullDir)
		if err != nil {
			t.Errorf("Failed to read directory %s: %v", dir, err)
			continue
		}

		if len(files) != len(expectedFiles) {
			t.Errorf("Directory %s: expected %d files, got %d", dir, len(expectedFiles), len(files))
		}
	}
}

func TestTransferManager_PriorityFiles(t *testing.T) {
	mgr := &Manager{
		config: &config.Config{
			Transfer: config.TransferConfig{
				PriorityPrefixes: []string{"1_", "urgent_"},
			},
		},
	}

	tests := []struct {
		filename string
		priority bool
	}{
		{"1_Project_Client_ACam_001.mp4", true},
		{"urgent_Project_Client_BCam_002.mp4", true},
		{"Project_Client_CCam_003.mp4", false},
		{"normal_file.mp4", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			isPriority := mgr.isPriorityFile(tt.filename)
			if isPriority != tt.priority {
				t.Errorf("Expected priority=%v for %s, got %v", tt.priority, tt.filename, isPriority)
			}
		})
	}
}

func TestTransferManager_Checksums(t *testing.T) {
	// Create temporary directories
	sourceDir, err := ioutil.TempDir("", "media-ingest-source-*")
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	destDir, err := ioutil.TempDir("", "media-ingest-dest-*")
	if err != nil {
		t.Fatalf("Failed to create dest dir: %v", err)
	}
	defer os.RemoveAll(destDir)

	logDir, err := ioutil.TempDir("", "media-ingest-logs-*")
	if err != nil {
		t.Fatalf("Failed to create log dir: %v", err)
	}
	defer os.RemoveAll(logDir)

	// Create a test file with known content
	testContent := []byte("This is test content for checksum verification")
	testFile := filepath.Join(sourceDir, "Test_Client_ACam_checksum.mp4")
	if err := ioutil.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create config with checksum verification enabled
	cfg := &config.Config{
		DestinationPath: destDir,
		Logging: config.LoggingConfig{
			ServerLogPath: logDir,
			LogToDevice:   false,
			LogLevel:      "debug",
		},
		Transfer: config.TransferConfig{
			MaxWorkers:      1,
			BufferSize:      1024,
			VerifyChecksums: true,
			MaxRetries:      3,
		},
		Parsing: config.ParsingConfig{
			Pattern:         "^([^_]+)_([^_]+)_(ACam|BCam|CCam)_(.+)$",
			FolderStructure: "{client}/{project}/{camera}",
			UnmatchedFolder: "Unsorted",
		},
		Performance: config.PerfConfig{
			ShowProgress:     false,
			ColoredOutput:    false,
		},
	}

	// Create logger and parser
	log, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	p, err := parser.NewParser(cfg)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Transfer the file
	mgr := NewManager(cfg, log, p)
	err = mgr.TransferFiles("test-device", []string{testFile})
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	// Verify the file was copied correctly
	destFile := filepath.Join(destDir, "Client", "Test", "ACam", "checksum.mp4")
	destContent, err := ioutil.ReadFile(destFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(destContent) != string(testContent) {
		t.Errorf("Content mismatch: expected %s, got %s", string(testContent), string(destContent))
	}
}
