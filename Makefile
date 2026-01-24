.PHONY: all clean build-arm64 build-arm build-all

# Default target
all: build-arm64

# Build for ARM64 (Raspberry Pi 4/5, CM5)
build-arm64:
	GOOS=linux GOARCH=arm64 go build -o nanoctl-linux-arm64

# Build for ARM 32-bit (Raspberry Pi 3 and earlier)
build-arm:
	GOOS=linux GOARCH=arm GOARM=7 go build -o nanoctl-linux-arm

# Build for both architectures
build-all: build-arm64 build-arm

# Clean build artifacts
clean:
	rm -f nanoctl-linux-arm64 nanoctl-linux-arm nanoctl

# For building on Linux directly
build:
	go build -o nanoctl
