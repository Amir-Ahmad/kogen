## Tests

Testing in this repository uses Testscript https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript, which is a copy of code used to test Go itself.

## Running tests

```
# Run local tests
go test -count=1 ./...

# Run local and remote tests
RUN_REMOTE=0 go test -count=1 ./...

# Run all tests and update any golden files
# Useful when writing new tests or making changes
UPDATE_GOLDEN=0 RUN_REMOTE=0 go test -count=1 ./...
```
