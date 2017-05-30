// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package main

import (
	"flag"

	"github.com/jtolds/golincs/mmm"
)

var (
	outputPath = flag.String("o", "", "output path")
	byCol      = flag.Bool("col", false,
		"if true, combine by adding columns instead of rows")
)

func equal(a, b []mmm.Ident) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}

func main() {
	flag.Parse()

	if *outputPath == "" {
		panic("output path (-o) required")
	}

	var handles []*mmm.Handle
	for _, path := range flag.Args() {
		fh, err := mmm.Open(path)
		if err != nil {
			panic(err)
		}
		defer fh.Close()
		handles = append(handles, fh)
	}

	rows, cols := 0, 0
	if len(handles) > 0 {
		rows, cols = handles[0].Rows(), handles[0].Cols()
		if *byCol {
			for _, handle := range handles[1:] {
				if !equal(handle.RowIds(), handles[0].RowIds()) {
					panic("row ids don't match")
				}
				cols += handle.Cols()
			}
		} else {
			for _, handle := range handles[1:] {
				if !equal(handle.ColIds(), handles[0].ColIds()) {
					panic("col ids don't match")
				}
				rows += handle.Rows()
			}
		}
	}

	out, err := mmm.Create(*outputPath, int64(rows), int64(cols))
	if err != nil {
		panic(err)
	}
	defer out.Close()

	if len(handles) > 0 {
		if *byCol {
			copy(out.RowIds(), handles[0].RowIds())
		} else {
			copy(out.ColIds(), handles[0].ColIds())
		}
	}

	if *byCol {
		offset := 0
		for _, handle := range handles {
			for i := 0; i < out.Rows(); i++ {
				copy(out.RowByIdx(i)[offset:offset+handle.Cols()], handle.RowByIdx(i))
			}
			copy(out.ColIds()[offset:offset+handle.Cols()], handle.ColIds())
			offset += handle.Cols()
		}
	} else {
		offset := 0
		for _, handle := range handles {
			for i := 0; i < handle.Rows(); i++ {
				copy(out.RowByIdx(offset+i), handle.RowByIdx(i))
			}
			copy(out.RowIds()[offset:offset+handle.Rows()], handle.RowIds())
			offset += handle.Rows()
		}
	}

	err = out.Close()
	if err != nil {
		panic(err)
	}
}
