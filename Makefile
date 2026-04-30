.PHONY: build dev docker test clean

build:
	go build -o nforge .

dev:
	gin --runtime gin

docker:
	docker build -t nfv2:latest .

test:
	go test ./...

clean:
	rm -f nforge
	rm -rf frontend/dist

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build
