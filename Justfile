# Print help
@help:
    just --list

@test:
    gotestsum -f testname ./... -- -count=1

@coverage:
    gotestsum -- -cover -coverprofile=c.out -count=1 ./...
    go tool cover -html=c.out

@test-all:
    RUN_REMOTE=0 gotestsum -f testname ./... -- -count=1
