// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package main

import (
	"flag"

	"github.com/jtolds/golincs/mmm"
)

func main() {
	flag.Parse()
	fh, err := mmm.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	for idx := 0; idx < fh.Rows(); idx++ {
		row := fh.RowByIdx(idx)
		for col := range row {
			row[col] = -row[col]
		}
	}

	err = fh.Close()
	if err != nil {
		panic(err)
	}
}
