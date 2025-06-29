# Print help
@help:
    just --list

@test:
    gotestsum -f testname ./... -- -count=1
