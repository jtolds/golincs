// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package main

import (
	"flag"
	"strconv"
	"strings"

	"github.com/jtolds/golincs/mmm"
)

var (
	rowsFlag = flag.String("rows", "", "comma-separated list of row ids. "+
		"will be removed, unless list starts with a '+'")
	colsFlag = flag.String("cols", "", "comma-separated list of col ids. "+
		"will be removed, unless list starts with a '+'")
	inputPath  = flag.String("i", "", "input path")
	outputPath = flag.String("o", "", "output path")
)

func main() {
	flag.Parse()

	if *inputPath == "" {
		panic("input path (-i) required")
	}
	if *outputPath == "" {
		panic("output path (-o) required")
	}

	var rowsInverted, colsInverted bool
	if strings.HasPrefix(*rowsFlag, "+") {
		*rowsFlag = strings.TrimPrefix(*rowsFlag, "+")
		rowsInverted = true
	}
	if strings.HasPrefix(*colsFlag, "+") {
		*colsFlag = strings.TrimPrefix(*colsFlag, "+")
		colsInverted = true
	}

	var rows, cols []uint32
	for _, part := range strings.Split(*rowsFlag, ",") {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			continue
		}
		id, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			panic(err)
		}
		rows = append(rows, uint32(id))
	}

	for _, part := range strings.Split(*colsFlag, ",") {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			continue
		}
		id, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			panic(err)
		}
		cols = append(cols, uint32(id))
	}

	err := mmm.Filter(*outputPath, *inputPath,
		rows, rowsInverted, cols, colsInverted)
	if err != nil {
		panic(err)
	}
}
