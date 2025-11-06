# Windows Test Script for Media Ingest Server

Write-Host "===============================================================" -ForegroundColor Cyan
Write-Host "     Media Ingest Server - Windows Test Suite                " -ForegroundColor Cyan
Write-Host "===============================================================" -ForegroundColor Cyan
Write-Host ""

$TEST_SOURCE = "C:\Temp\media-ingest-test-source"
$TEST_DEST = "C:\Temp\media-ingest-test-dest"
$TEST_CONFIG = "C:\Temp\media-ingest-test-config.yaml"
$TEST_LOGS = "C:\Temp\media-ingest-test-logs"

function Cleanup {
    Write-Host ""
    Write-Host "Cleaning up test directories..." -ForegroundColor Yellow
    if (Test-Path $TEST_SOURCE) { Remove-Item -Recurse -Force $TEST_SOURCE }
    if (Test-Path $TEST_DEST) { Remove-Item -Recurse -Force $TEST_DEST }
    if (Test-Path $TEST_LOGS) { Remove-Item -Recurse -Force $TEST_LOGS }
    if (Test-Path $TEST_CONFIG) { Remove-Item -Force $TEST_CONFIG }
    Write-Host "[OK] Cleanup complete" -ForegroundColor Green
}

Write-Host "Setting up test environment..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path $TEST_SOURCE | Out-Null
New-Item -ItemType Directory -Force -Path $TEST_DEST | Out-Null
New-Item -ItemType Directory -Force -Path $TEST_LOGS | Out-Null

Write-Host "Creating test files..." -ForegroundColor Yellow
"test data" | Out-File "$TEST_SOURCE\1_BrandVideo_Nike_ACam_001.mp4"
"test data" | Out-File "$TEST_SOURCE\1_BrandVideo_Nike_BCam_001.mp4"
"test data" | Out-File "$TEST_SOURCE\ProductShoot_Adidas_ACam_042.mov"
"test data" | Out-File "$TEST_SOURCE\ProductShoot_Adidas_CCam_043.mxf"
"test data" | Out-File "$TEST_SOURCE\Interview_Tesla_BCam_Take5.mp4"
"test data" | Out-File "$TEST_SOURCE\random_video.mp4"
"test data" | Out-File "$TEST_SOURCE\no_pattern.mov"

$largeFile = "$TEST_SOURCE\1_LargeProject_Client_ACam_large.mp4"
$size = 50MB
$bytes = New-Object byte[] $size
(New-Object Random).NextBytes($bytes)
[IO.File]::WriteAllBytes($largeFile, $bytes)

Write-Host "[OK] Created test files" -ForegroundColor Green
Write-Host ""

Write-Host "Creating test configuration..." -ForegroundColor Yellow
$configContent = @"
destination_path: "$($TEST_DEST.Replace('\', '\\'))"

auto_mount:
  mount_base: ""
  enabled: false

logging:
  server_log_path: "$($TEST_LOGS.Replace('\', '\\'))"
  log_to_device: false
  retention_days: 7
  log_level: "debug"

transfer:
  max_workers: 4
  buffer_size: 1048576
  verify_checksums: true
  max_retries: 3
  priority_prefixes:
    - "1_"
    - "priority_"

parsing:
  pattern: "^([^_]+)_([^_]+)_(ACam|BCam|CCam)_(.+)$"
  folder_structure: "{client}/{project}/{camera}"
  unmatched_folder: "Unsorted"

email:
  enabled: false

device_detection:
  enabled: true
  min_size_bytes: 1024
  allowed_filesystems:
    - "NTFS"
    - "FAT32"
    - "exFAT"
  exclude_patterns:
    - "C:\\"

performance:
  show_progress: true
  progress_interval: 1
  colored_output: true
"@

$configContent | Out-File -FilePath $TEST_CONFIG -Encoding UTF8
Write-Host "[OK] Created test configuration" -ForegroundColor Green
Write-Host ""

Write-Host "Running unit tests..." -ForegroundColor Cyan
go test ./internal/parser/... -v
if ($LASTEXITCODE -eq 0) {
    Write-Host "[PASS] Unit tests passed" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Unit tests failed" -ForegroundColor Red
    Cleanup
    exit 1
}
Write-Host ""

Write-Host "Running integration tests..." -ForegroundColor Cyan
go test ./internal/transfer/... -v
if ($LASTEXITCODE -eq 0) {
    Write-Host "[PASS] Integration tests passed" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Integration tests failed" -ForegroundColor Red
    Cleanup
    exit 1
}
Write-Host ""

Write-Host "Building application..." -ForegroundColor Cyan
go build -o media-ingest-test.exe ./cmd/media-ingest
if ($LASTEXITCODE -eq 0) {
    Write-Host "[PASS] Build successful" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Build failed" -ForegroundColor Red
    Cleanup
    exit 1
}
Write-Host ""

Write-Host "===============================================================" -ForegroundColor Green
Write-Host "              All Tests Passed!                               " -ForegroundColor Green
Write-Host "===============================================================" -ForegroundColor Green
Write-Host ""

Write-Host "Test directories:" -ForegroundColor Cyan
Write-Host "  Source: $TEST_SOURCE"
Write-Host "  Destination: $TEST_DEST"
Write-Host "  Config: $TEST_CONFIG"
Write-Host "  Logs: $TEST_LOGS"
Write-Host ""

Write-Host "To manually test:" -ForegroundColor Cyan
Write-Host "  .\media-ingest-test.exe -config $TEST_CONFIG"
Write-Host ""

Write-Host "Press any key to cleanup or Ctrl+C to keep test files..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
Cleanup
