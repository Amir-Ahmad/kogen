## Tests

Testing in this repository uses Testscript https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript, which is a copy of code used to test Go itself.

## Running tests

```
go test -count=1 ./...

# Update golden when writing new tests or making changes
UPDATE_GOLDEN=0 go test -count=1 ./...
```
