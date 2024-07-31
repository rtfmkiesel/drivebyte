.PHONY: make
make:
	go mod tidy
	go build -ldflags="-s -w" -o drivebyte .

.PHONY: install
install:
	go mod tidy
	go install -ldflags="-s -w" .