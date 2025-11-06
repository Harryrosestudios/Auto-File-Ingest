.PHONY: build install clean test run dev

# Build the application
build:
	go mod download
	go build -o media-ingest ./cmd/media-ingest

# Install system-wide (requires root)
install: build
	sudo install -m 755 media-ingest /usr/local/bin/media-ingest
	sudo mkdir -p /etc/media-ingest /var/log/media-ingest /mnt/ingest /mnt/storage/media
	sudo cp config.example.yaml /etc/media-ingest/config.yaml
	sudo cp systemd/media-ingest.service /etc/systemd/system/
	sudo cp udev/99-media-ingest.rules /etc/udev/rules.d/
	sudo systemctl daemon-reload
	sudo udevadm control --reload-rules
	sudo udevadm trigger
	@echo "Installation complete! Edit /etc/media-ingest/config.yaml and run 'sudo systemctl start media-ingest'"

# Clean build artifacts
clean:
	rm -f media-ingest
	go clean

# Run tests
test:
	go test -v ./...

# Run development version
dev: build
	sudo ./media-ingest -config config.example.yaml

# Run with race detection
race:
	go run -race ./cmd/media-ingest -config config.example.yaml

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Show service status
status:
	sudo systemctl status media-ingest

# View logs
logs:
	sudo journalctl -u media-ingest -f

# Uninstall
uninstall:
	sudo systemctl stop media-ingest || true
	sudo systemctl disable media-ingest || true
	sudo rm -f /usr/local/bin/media-ingest
	sudo rm -f /etc/systemd/system/media-ingest.service
	sudo rm -f /etc/udev/rules.d/99-media-ingest.rules
	sudo systemctl daemon-reload
	sudo udevadm control --reload-rules
	@echo "Uninstalled. Config and logs remain in /etc/media-ingest and /var/log/media-ingest"
