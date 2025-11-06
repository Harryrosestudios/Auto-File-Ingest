package parser

import (
	"path/filepath"
	"testing"

	"github.com/autofileingest/internal/config"
)

func TestParser_Parse(t *testing.T) {
	cfg := &config.Config{
		Parsing: config.ParsingConfig{
			Pattern:         "^([^_]+)_([^_]+)_(ACam|BCam|CCam)_(.+)$",
			FolderStructure: "{client}/{project}/{camera}",
			UnmatchedFolder: "Unsorted",
		},
		DestinationPath: "/mnt/storage",
	}

	parser, err := NewParser(cfg)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		name           string
		filename       string
		expectedMatch  bool
		expectedClient string
		expectedProject string
		expectedCamera string
	}{
		{
			name:           "Valid ACam file",
			filename:       "/path/to/BrandVideo_Nike_ACam_001.mp4",
			expectedMatch:  true,
			expectedClient: "Nike",
			expectedProject: "BrandVideo",
			expectedCamera: "ACam",
		},
		{
			name:           "Valid BCam file with priority prefix",
			filename:       "/path/to/ProductShoot_Adidas_BCam_042.mov",
			expectedMatch:  true,
			expectedClient: "Adidas",
			expectedProject: "ProductShoot",
			expectedCamera: "BCam",
		},
		{
			name:           "Valid CCam file",
			filename:       "Interview_Tesla_CCam_Take5.mxf",
			expectedMatch:  true,
			expectedClient: "Tesla",
			expectedProject: "Interview",
			expectedCamera: "CCam",
		},
		{
			name:           "Invalid file - no pattern match",
			filename:       "random_video.mp4",
			expectedMatch:  false,
		},
		{
			name:           "Invalid file - wrong camera",
			filename:       "Project_Client_DCam_001.mp4",
			expectedMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := parser.Parse(tt.filename)

			if info.Matched != tt.expectedMatch {
				t.Errorf("Expected matched=%v, got %v", tt.expectedMatch, info.Matched)
			}

			if tt.expectedMatch {
				if info.Client != tt.expectedClient {
					t.Errorf("Expected client=%s, got %s", tt.expectedClient, info.Client)
				}
				if info.ProjectName != tt.expectedProject {
					t.Errorf("Expected project=%s, got %s", tt.expectedProject, info.ProjectName)
				}
				if info.Camera != tt.expectedCamera {
					t.Errorf("Expected camera=%s, got %s", tt.expectedCamera, info.Camera)
				}
			}
		})
	}
}

func TestParser_GetDestinationPath(t *testing.T) {
	cfg := &config.Config{
		Parsing: config.ParsingConfig{
			Pattern:         "^([^_]+)_([^_]+)_(ACam|BCam|CCam)_(.+)$",
			FolderStructure: "{client}/{project}/{camera}",
			UnmatchedFolder: "Unsorted",
		},
		DestinationPath: "/mnt/storage",
	}

	parser, err := NewParser(cfg)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		name         string
		filename     string
		expectedPath string
	}{
		{
			name:         "Matched file",
			filename:     "BrandVideo_Nike_ACam_001.mp4",
			expectedPath: "Nike/BrandVideo/ACam",
		},
		{
			name:         "Unmatched file",
			filename:     "random.mp4",
			expectedPath: "Unsorted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := parser.Parse(tt.filename)
			path := parser.GetDestinationPath(info)

			// Normalize path separators for cross-platform compatibility
			normalizedPath := filepath.ToSlash(path)
			expectedFullPath := "/mnt/storage/" + tt.expectedPath
			
			if normalizedPath != expectedFullPath {
				t.Errorf("Expected path=%s, got %s", expectedFullPath, normalizedPath)
			}
		})
	}
}
