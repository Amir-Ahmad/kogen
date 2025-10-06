# Print help
@help:
    just --list

# Run unit + local e2e tests
@test:
    gotestsum --hide-summary=skipped -f testname ./... -- -count=1

# Run all tests with coverage report
@coverage:
    RUN_REMOTE=0 gotestsum -- -cover -coverprofile=c.out -count=1 -coverpkg=./... ./...
    go tool cover -html=c.out

# Run all tests
@test-all:
    RUN_REMOTE=0 gotestsum -f testname ./... -- -count=1

# Build binary for amd64 linux
@build-linux:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/kogen main.go

# Build docker image for argo plugin
@docker-build:
    docker build -t ghcr.io/amir-ahmad/kogen:scratch -f argo-plugin/Dockerfile .
