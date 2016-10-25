# golincs

This repository contains packages to assist with the LINCS project written in
Go.

See the documentation at https://godoc.org/github.com/jtolds/golincs/

## Example

Currently, just the characteristic direction signature method is implemented.
You can try out the Go implementation by running:

```
$ cd $(mktemp -d)
/tmp/tmp.rqR3ZNiIbn $ GOPATH=$(pwd) go get github.com/jtolds/golincs/cmd/cds
/tmp/tmp.rqR3ZNiIbn $ bin/cds --input ~/path/to/input.txt
```

This package depends heavily on https://github.com/gonum/matrix, which itself
may require having BLAS and/or LAPACK installed. Installation instructions for
the gonum bindings for those dependencies can be found at:

 * https://github.com/gonum/blas
 * https://github.com/gonum/lapack
