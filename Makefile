APP := urlshorty

.PHONY: run build test fmt lint

run:
	go run ./cmd/$(APP)

build:
	go build -ldflags="-s -w" -o bin/$(APP) ./cmd/$(APP)

test:
	go test ./...

fmt:
	gofmt -s -w .
	@command -v gofumpt >/dev/null 2>&1 && gofumpt -l -w . || true

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || \
	echo "golangci-lint not installed; skipping"
