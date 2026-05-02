.PHONY: build build-linux build-windows dev docker test clean frontend-install frontend-build help

BINARY = nforge
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -X github.com/nnlgsakib/nodeforge/cmd/nforge.version=$(VERSION) -s -w
BUILD_FLAGS = -trimpath
GOARCH := amd64

# Detect OS
ifneq ($(OS),Windows_NT)
  UNAME_S := $(shell uname -s)
  ifeq ($(UNAME_S),Linux)
    TARGET_OS := linux
    TARGET_EXT :=
  endif
  ifeq ($(UNAME_S),Darwin)
    TARGET_OS := darwin
    TARGET_EXT :=
  endif
else
  TARGET_OS := windows
  TARGET_EXT := .exe
endif

# Allow overriding target OS via make GOOS=linux or GOOS=windows
GOOS ?= $(TARGET_OS)
VALID_GOOS := linux darwin windows
ifeq ($(filter $(GOOS),$(VALID_GOOS)),)
  $(error Invalid GOOS=$(GOOS). Valid values: linux, darwin, windows)
endif
ifeq ($(GOOS),windows)
  TARGET_EXT := .exe
else
  TARGET_EXT :=
endif

BINARY_PATH = $(BINARY)$(TARGET_EXT)

# Export GOOS and GOARCH so go build sees them without inline VAR=val syntax
export GOOS
export GOARCH

build: frontend-build
	@echo "Building for $(GOOS)/$(GOARCH)..."
	go build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) main.go
	@echo "Built: $(BINARY_PATH)"

build-linux: export GOOS := linux
build-linux: frontend-build
	@echo "Cross-compiling for Linux..."
	go build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) main.go
	@echo "Built: $(BINARY)"

build-windows: export GOOS := windows
build-windows: frontend-build
	@echo "Cross-compiling for Windows..."
	go build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY).exe main.go
	@echo "Built: $(BINARY).exe"

dev:
	cd frontend && npm run dev

docker:
	docker build -t nfv2:latest .

test:
	go test ./...

clean:
	rm -f $(BINARY)
	rm -f $(BINARY).exe
	rm -rf frontend/dist

frontend-install:
	cd frontend && npm install

frontend-build: frontend-install
	cd frontend && npm run build

help:
	@echo "Available targets:"
	@echo "  build          - Build for detected OS (or set GOOS=linux|darwin|windows)"
	@echo "  build-linux    - Cross-compile for Linux"
	@echo "  build-windows  - Cross-compile for Windows"
	@echo "  dev            - Start frontend dev server"
	@echo "  test           - Run Go tests"
	@echo "  clean          - Remove build artifacts"
	@echo "  frontend-install - Install frontend dependencies"
	@echo "  frontend-build - Build frontend for production"
	@echo "  docker         - Build Docker image"
	@echo ""
	@echo "Examples:"
	@echo "  make build              # Build for current OS"
	@echo "  make build GOOS=linux   # Build for Linux"
	@echo "  make build GOOS=darwin  # Build for macOS"
	@echo "  make build GOOS=windows # Build for Windows"
	@echo ""
	@echo "Valid GOOS values: linux, darwin, windows"
