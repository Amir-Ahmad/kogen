# Print help
@help:
    just --list

@test:
    gotestsum --hide-summary=skipped -f testname ./... -- -count=1

@coverage:
    RUN_REMOTE=0 gotestsum -- -cover -coverprofile=c.out -count=1 -coverpkg=./... ./...
    go tool cover -html=c.out

@test-all:
    RUN_REMOTE=0 gotestsum -f testname ./... -- -count=1
