.PHONY: build test test-cover clean deps generate vet lint

build:
	go build -o strava-mcp .

test:
	go test -timeout 30s ./...

test-cover:
	go test -timeout 30s -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

clean:
	rm -f strava-mcp coverage.out

deps:
	go mod tidy

generate:
	sqlc generate

vet:
	go vet ./...

lint: vet
	@echo "Running go vet... done"
