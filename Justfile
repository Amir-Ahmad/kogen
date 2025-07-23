# Print help
@help:
    just --list

@test:
    gotestsum -f testname ./... -- -count=1

@test-all:
    RUN_REMOTE=0 gotestsum -f testname ./... -- -count=1
