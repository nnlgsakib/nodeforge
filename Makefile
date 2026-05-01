.PHONY: build dev docker test clean frontend-install frontend-build

BINARY = nforge
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -X github.com/nlg/nfv2/cmd/nforge.version=$(VERSION)

build: frontend-build
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) main.go

dev:
	cd frontend && npm run dev

docker:
	docker build -t nfv2:latest .

test:
	go test ./...

clean:
	rm -f $(BINARY)

frontend-install:
	cd frontend && npm install

frontend-build:
	cd frontend && npm run build
