// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package main

import (
	"flag"
	"fmt"

	"github.com/jtolds/golincs/mmm"
)

func main() {
	flag.Parse()
	for i, path := range flag.Args() {
		if i != 0 {
			_, err := fmt.Println()
			if err != nil {
				panic(err)
			}
		}
		fh, err := mmm.Open(path)
		if err != nil {
			panic(err)
		}
		defer fh.Close()

		_, err = fmt.Printf("Rows: %d\nCols: %d\n", fh.Rows(), fh.Cols())
		if err != nil {
			panic(err)
		}
	}
}
