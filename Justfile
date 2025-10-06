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

# Build binary
@build:
    goreleaser build --snapshot --clean --single-target

# Build binary and docker images
build-all:
    goreleaser release --snapshot --clean --skip=archive
