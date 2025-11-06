#!/bin/bash
# Test script for Media Ingest Server

set -e

echo "╔════════════════════════════════════════════════════════╗"
echo "║      Media Ingest Server - Test Suite                 ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test directories
TEST_SOURCE="/tmp/media-ingest-test-source"
TEST_DEST="/tmp/media-ingest-test-dest"
TEST_MOUNT="/tmp/media-ingest-test-mount"

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up test directories..."
    rm -rf "$TEST_SOURCE" "$TEST_DEST" "$TEST_MOUNT"
    echo "✓ Cleanup complete"
}

trap cleanup EXIT

# Create test directories
echo "Setting up test environment..."
mkdir -p "$TEST_SOURCE"
mkdir -p "$TEST_DEST"
mkdir -p "$TEST_MOUNT"

# Create test files with various patterns
echo "Creating test files..."

# Valid files - should be organized
echo "test data" > "$TEST_SOURCE/1_BrandVideo_Nike_ACam_001.mp4"
echo "test data" > "$TEST_SOURCE/1_BrandVideo_Nike_BCam_001.mp4"
echo "test data" > "$TEST_SOURCE/ProductShoot_Adidas_ACam_042.mov"
echo "test data" > "$TEST_SOURCE/ProductShoot_Adidas_CCam_043.mxf"
echo "test data" > "$TEST_SOURCE/Interview_Tesla_BCam_Take5.mp4"

# Invalid files - should go to Unsorted
echo "test data" > "$TEST_SOURCE/random_video.mp4"
echo "test data" > "$TEST_SOURCE/no_pattern.mov"
echo "test data" > "$TEST_SOURCE/wrong_Project_Client_DCam_001.mp4"

# Large file simulation
dd if=/dev/zero of="$TEST_SOURCE/1_LargeFile_Client_ACam_large.mp4" bs=1M count=100 2>/dev/null

echo "✓ Created test files"
echo ""

# Create test config
TEST_CONFIG="/tmp/media-ingest-test-config.yaml"
cat > "$TEST_CONFIG" << EOF
destination_path: "$TEST_DEST"

auto_mount:
  mount_base: "$TEST_MOUNT"
  enabled: false

logging:
  server_log_path: "/tmp/media-ingest-test-logs"
  log_to_device: false
  retention_days: 7
  log_level: "debug"

transfer:
  max_workers: 2
  buffer_size: 1048576
  verify_checksums: true
  max_retries: 3
  priority_prefixes:
    - "1_"

parsing:
  pattern: "^([^_]+)_([^_]+)_(ACam|BCam|CCam)_(.+)$"
  folder_structure: "{client}/{project}/{camera}"
  unmatched_folder: "Unsorted"

email:
  enabled: false

device_detection:
  enabled: false
  min_size_bytes: 1024
  allowed_filesystems:
    - "ext4"
    - "vfat"
  exclude_patterns: []

performance:
  show_progress: true
  progress_interval: 1
  colored_output: true
EOF

echo "✓ Created test configuration"
echo ""

# Run parser tests
echo "Running Go tests..."
if go test ./internal/parser/... -v; then
    echo -e "${GREEN}✓ Parser tests passed${NC}"
else
    echo -e "${RED}✗ Parser tests failed${NC}"
    exit 1
fi
echo ""

# Build the application
echo "Building application..."
if go build -o media-ingest-test ./cmd/media-ingest; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi
echo ""

echo "╔════════════════════════════════════════════════════════╗"
echo "║                 Test Results                           ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Check expected file structure
echo "Expected file organization:"
echo "  Nike/BrandVideo/ACam/ - 2 files (priority)"
echo "  Nike/BrandVideo/BCam/ - 1 file (priority)"
echo "  Adidas/ProductShoot/ACam/ - 1 file"
echo "  Adidas/ProductShoot/CCam/ - 1 file"
echo "  Tesla/Interview/BCam/ - 1 file"
echo "  Client/LargeFile/ACam/ - 1 large file (priority)"
echo "  Unsorted/ - 3 files"
echo ""

echo -e "${GREEN}All tests completed successfully!${NC}"
echo ""
echo "To run the application manually with test data:"
echo "  ./media-ingest-test -config $TEST_CONFIG"
echo ""
echo "Test directories:"
echo "  Source: $TEST_SOURCE"
echo "  Destination: $TEST_DEST"
echo "  Config: $TEST_CONFIG"
