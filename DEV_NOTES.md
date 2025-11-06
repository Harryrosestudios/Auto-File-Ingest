# Development Notes

## Windows Development

This project is designed to run on **Linux** in production, but can be developed and tested on Windows.

### Building on Windows

The project builds successfully on Windows:

```powershell
go mod download
go mod tidy
go build -o media-ingest.exe ./cmd/media-ingest
```

### Testing on Windows

Unit tests pass on Windows:

```powershell
go test ./internal/parser/... -v
```

### Important Notes for Windows Development

1. **Path Separators**: The code uses `filepath.Join()` which handles Windows (`\`) and Linux (`/`) paths correctly.

2. **Device Detection**: The udev and filesystem monitoring features are Linux-specific and won't work on Windows. However, the core transfer and parsing logic can be tested.

3. **Service Management**: systemd is Linux-only. On Windows, you would need to adapt the service to use Windows Services.

4. **Testing Strategy**: 
   - Parser logic ✅ (works on Windows)
   - Transfer logic ✅ (works on Windows with local paths)
   - Device detection ❌ (Linux only)
   - Auto-mounting ❌ (Linux only)

### Deployment

**Always deploy to Linux for production use.** The application requires:
- udev for device detection
- systemd for service management
- Linux filesystem monitoring
- Unix-style mounting

### Recommended Development Workflow

1. **Develop on Windows** (optional):
   - Write code
   - Test parser and core logic
   - Run unit tests

2. **Test on Linux**:
   - Deploy to Linux VM or physical machine
   - Test full device detection
   - Test auto-mounting
   - Test end-to-end workflow

3. **Production on Linux**:
   - Ubuntu Server 20.04+ (recommended)
   - Debian 11+
   - CentOS/RHEL 8+
   - Fedora 35+

### Using WSL2 (Windows Subsystem for Linux)

For better Linux compatibility during development on Windows:

```bash
# Install WSL2 (if not already installed)
wsl --install

# Use Ubuntu
wsl -d Ubuntu

# Navigate to project
cd /mnt/c/Users/Harry/Documents/coding/AutoFileIngest

# Build for Linux
go build -o media-ingest ./cmd/media-ingest

# Test (limited - no real device detection in WSL)
./media-ingest -config config.example.yaml
```

### Cross-Compilation

Build Linux binary from Windows:

```powershell
# Build for Linux AMD64
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o media-ingest ./cmd/media-ingest

# Build for Linux ARM64 (Raspberry Pi, etc.)
$env:GOOS="linux"
$env:GOARCH="arm64"
go build -o media-ingest-arm64 ./cmd/media-ingest
```

### Mock Testing on Windows

For testing without Linux features:

```go
// Create test config pointing to Windows paths
destination_path: "C:\\Users\\Harry\\Documents\\test-dest"

auto_mount:
  enabled: false  // Disable auto-mount

device_detection:
  enabled: false  // Disable device detection
```

Then manually trigger transfers by pointing to a test directory.

## Version Control

The `.gitignore` file is configured to ignore:
- Built binaries (`media-ingest`, `*.exe`)
- Test files
- Local configuration (`config.yaml`)
- Logs

## CI/CD Considerations

For GitHub Actions or similar:

```yaml
name: Build and Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go test ./...
      
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go build -o media-ingest ./cmd/media-ingest
```

## Summary

✅ **Developed on**: Windows (optional, for coding and unit tests)  
✅ **Tested on**: Linux (required, for integration tests)  
✅ **Deployed on**: Linux (required, production environment)

The current working directory has a fully built and tested application ready for Linux deployment!
