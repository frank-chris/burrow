BINARY = burrow
VERSION ?= dev

.PHONY: build install clean snapshot

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) .

install:
	go install -ldflags "-X main.version=$(VERSION)" .

clean:
	rm -f $(BINARY) $(BINARY).exe
	rm -rf dist/

snapshot:
	goreleaser release --snapshot --clean
