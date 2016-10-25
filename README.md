# golincs

This repository contains packages to assist with the LINCS project written in
Go.

See the documentation at https://godoc.org/github.com/jtolds/golincs/

## Example

Currently, just the characteristic direction signature method is implemented.
You can try out the Go implementation by running:

```
$ GOPATH=$(pwd) go get github.com/jtolds/golincs/cmd/cds
$ bin/cds --input path-to-input.txt
```
