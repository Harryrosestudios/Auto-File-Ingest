package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	DestinationPath string          `yaml:"destination_path"`
	AutoMount       AutoMountConfig `yaml:"auto_mount"`
	Logging         LoggingConfig   `yaml:"logging"`
	Transfer        TransferConfig  `yaml:"transfer"`
	Parsing         ParsingConfig   `yaml:"parsing"`
	Email           EmailConfig     `yaml:"email"`
	DeviceDetection DeviceConfig    `yaml:"device_detection"`
	Performance     PerfConfig      `yaml:"performance"`
}

type AutoMountConfig struct {
	MountBase string `yaml:"mount_base"`
	Enabled   bool   `yaml:"enabled"`
}

type LoggingConfig struct {
	ServerLogPath string `yaml:"server_log_path"`
	LogToDevice   bool   `yaml:"log_to_device"`
	RetentionDays int    `yaml:"retention_days"`
	LogLevel      string `yaml:"log_level"`
}

type TransferConfig struct {
	MaxWorkers       int      `yaml:"max_workers"`
	BufferSize       int      `yaml:"buffer_size"`
	VerifyChecksums  bool     `yaml:"verify_checksums"`
	MaxRetries       int      `yaml:"max_retries"`
	PriorityPrefixes []string `yaml:"priority_prefixes"`
}

type ParsingConfig struct {
	Pattern         string `yaml:"pattern"`
	FolderStructure string `yaml:"folder_structure"`
	UnmatchedFolder string `yaml:"unmatched_folder"`
}

type EmailConfig struct {
	Enabled    bool     `yaml:"enabled"`
	SMTPHost   string   `yaml:"smtp_host"`
	SMTPPort   int      `yaml:"smtp_port"`
	UseTLS     bool     `yaml:"use_tls"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	From       string   `yaml:"from"`
	To         []string `yaml:"to"`
	Subject    string   `yaml:"subject"`
	AttachLog  bool     `yaml:"attach_log"`
}

type DeviceConfig struct {
	Enabled            bool     `yaml:"enabled"`
	MinSizeBytes       int64    `yaml:"min_size_bytes"`
	AllowedFilesystems []string `yaml:"allowed_filesystems"`
	ExcludePatterns    []string `yaml:"exclude_patterns"`
}

type PerfConfig struct {
	ShowProgress     bool `yaml:"show_progress"`
	ProgressInterval int  `yaml:"progress_interval"`
	ColoredOutput    bool `yaml:"colored_output"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.DestinationPath == "" {
		return fmt.Errorf("destination_path is required")
	}

	if c.Transfer.MaxWorkers < 1 {
		c.Transfer.MaxWorkers = 1
	}

	if c.Transfer.BufferSize < 1024 {
		c.Transfer.BufferSize = 1048576 // 1MB default
	}

	if c.Parsing.Pattern == "" {
		return fmt.Errorf("parsing.pattern is required")
	}

	if c.Email.Enabled {
		if c.Email.SMTPHost == "" || c.Email.SMTPPort == 0 {
			return fmt.Errorf("email is enabled but SMTP settings are incomplete")
		}
		if len(c.Email.To) == 0 {
			return fmt.Errorf("email is enabled but no recipients specified")
		}
	}

	return nil
}
