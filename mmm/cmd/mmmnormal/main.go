// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package main

import (
	"flag"
	"math"

	"github.com/jtolds/golincs/mmm"
)

func normalize(vector []float32) {
	var squared_sum float64
	for _, val := range vector {
		squared_sum += float64(val) * float64(val)
	}
	dist := math.Sqrt(squared_sum)
	for i := range vector {
		vector[i] = float32(float64(vector[i]) / dist)
	}
}

func main() {
	flag.Parse()
	fh, err := mmm.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	for idx := 0; idx < fh.Rows(); idx++ {
		normalize(fh.RowByIdx(idx))
	}

	err = fh.Close()
	if err != nil {
		panic(err)
	}
}
