.PHONY: build test lint fmt vet clean ci fmt-check

build:
	go build -o oqx ./cmd/oqx

test:
	go test -race ./...

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f oqx coverage.txt coverage.html

ci: fmt vet test lint fmt-check

fmt-check:
	@test -z "$$(go fmt ./...)" || (echo "Files need formatting"; exit 1)
