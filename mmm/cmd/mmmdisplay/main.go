// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package main

import (
	"flag"
	"fmt"

	"github.com/jtolds/golincs/mmm"
)

func must(n int, err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		panic("expecting exactly one argument")
	}
	fh, err := mmm.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	for _, id := range fh.ColIds() {
		must(fmt.Printf("\t%d", id))
	}
	must(fmt.Println())

	for idx := 0; idx < fh.Rows(); idx++ {
		must(fmt.Printf("%d", fh.RowIds()[idx]))
		for _, val := range fh.RowByIdx(idx) {
			must(fmt.Printf("\t%f", val))
		}
		must(fmt.Println())
	}
}
