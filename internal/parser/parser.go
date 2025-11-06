package parser

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/autofileingest/internal/config"
)

// FileInfo represents parsed file information
type FileInfo struct {
	OriginalPath string
	FileName     string
	ProjectName  string
	Client       string
	Camera       string
	ClipNumber   string
	Extension    string
	Matched      bool
}

// Parser handles filename parsing
type Parser struct {
	pattern *regexp.Regexp
	config  *config.Config
}

// NewParser creates a new parser instance
func NewParser(cfg *config.Config) (*Parser, error) {
	pattern, err := regexp.Compile(cfg.Parsing.Pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid parsing pattern: %w", err)
	}

	return &Parser{
		pattern: pattern,
		config:  cfg,
	}, nil
}

// Parse extracts information from a filename
func (p *Parser) Parse(filePath string) *FileInfo {
	fileName := filepath.Base(filePath)
	
	info := &FileInfo{
		OriginalPath: filePath,
		FileName:     fileName,
		Extension:    filepath.Ext(fileName),
	}

	// Remove extension for parsing
	nameWithoutExt := strings.TrimSuffix(fileName, info.Extension)

	// Try to match pattern
	matches := p.pattern.FindStringSubmatch(nameWithoutExt)
	if len(matches) == 5 {
		info.ProjectName = matches[1]
		info.Client = matches[2]
		info.Camera = matches[3]
		info.ClipNumber = matches[4]
		info.Matched = true
	} else {
		info.Matched = false
	}

	return info
}

// GetDestinationPath returns the organized destination path for a file
func (p *Parser) GetDestinationPath(info *FileInfo) string {
	basePath := p.config.DestinationPath

	if !info.Matched {
		// Files that don't match go to unsorted folder
		return filepath.Join(basePath, p.config.Parsing.UnmatchedFolder)
	}

	// Build path from folder structure template
	structure := p.config.Parsing.FolderStructure
	structure = strings.ReplaceAll(structure, "{client}", info.Client)
	structure = strings.ReplaceAll(structure, "{project}", info.ProjectName)
	structure = strings.ReplaceAll(structure, "{camera}", info.Camera)

	return filepath.Join(basePath, structure)
}

// GetFullDestinationPath returns the complete destination path including filename
func (p *Parser) GetFullDestinationPath(info *FileInfo) string {
	destDir := p.GetDestinationPath(info)
	
	if info.Matched {
		fileName := fmt.Sprintf("%s%s", info.ClipNumber, info.Extension)
		return filepath.Join(destDir, fileName)
	}
	
	return filepath.Join(destDir, info.FileName)
}

// GetUniqueDestinationPath ensures the destination path is unique by adding version numbers
func (p *Parser) GetUniqueDestinationPath(info *FileInfo) (string, error) {
	destPath := p.GetFullDestinationPath(info)
	
	// Check if file exists
	if _, err := filepath.Glob(destPath); err == nil {
		// File doesn't exist, use as is
		return destPath, nil
	}

	// File exists, add version number
	dir := filepath.Dir(destPath)
	ext := filepath.Ext(destPath)
	nameWithoutExt := strings.TrimSuffix(filepath.Base(destPath), ext)

	version := 2
	for {
		versionedName := fmt.Sprintf("%s_v%d%s", nameWithoutExt, version, ext)
		versionedPath := filepath.Join(dir, versionedName)
		
		if _, err := filepath.Glob(versionedPath); err == nil {
			// This version doesn't exist
			return versionedPath, nil
		}
		
		version++
		if version > 1000 {
			return "", fmt.Errorf("too many versions of file: %s", destPath)
		}
	}
}
