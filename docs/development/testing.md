# Testing

Traditional unit tests for `cani` are in the `*_test.go` files.  Run `go test ./...` to run those tests.

Shellspec tests are used for integration-style tests, which test the usability of the CLI interface.  Run `make shellspec` to run those tests.
