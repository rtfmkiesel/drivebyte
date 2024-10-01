.PHONY: make
make:
	go mod tidy
	go build -ldflags="-s -w" .

.PHONY: install
install:
	go mod tidy
	go install -ldflags="-s -w" .

.PHONY: lint
lint:
	gosec -quiet ./...
	deadcode ./...
	goimports-reviser -format ./...
	golangci-lint run