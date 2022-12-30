lint:
	golangci-lint run

release:
	goreleaser build --snapshot --single-target --rm-dist

build:
	go build

test:
	go test -race ./...

cover:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
