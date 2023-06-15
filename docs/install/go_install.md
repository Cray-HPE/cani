# Go Install

1. Install Go.
1. Clone the cani repo: https://github.com/Cray-HPE/cani.git
1. Run or build the binary

```shell
go run main.go # run the app directly using go
make bin # use the makefile to make the binary with support for the version subcommand
go build -o bin/cani # quickly build without support for the version subcommand
```
