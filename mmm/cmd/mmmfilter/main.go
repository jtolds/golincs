// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package main

import (
	"bufio"
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/jtolds/golincs/mmm"
)

var (
	inputPath = flag.String(
		"i", "", "input path")
	outputPath = flag.String(
		"o", "", "output path")
	rowsFlag = flag.String(
		"rows", "", "comma-separated list of row ids.")
	colsFlag = flag.String(
		"cols", "", "comma-separated list of col ids.")
	rowsPathFlag = flag.String(
		"rows_path", "", "path to newline-separated list of row ids")
	colsPathFlag = flag.String(
		"cols_path", "", "path to newline-separated list of col ids")
	rowsInverted = flag.Bool(
		"row_keep", false, "if true, keep the rows, instead of removing them")
	colsInverted = flag.Bool(
		"col_keep", false, "if true, keep the columns, instead of removing them")
)

func getIds(flagval, path string) []uint32 {
	var ids []uint32
	add := func(part string) {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			return
		}
		id, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			panic(err)
		}
		ids = append(ids, uint32(id))
	}

	for _, part := range strings.Split(flagval, ",") {
		add(part)
	}

	if path != "" {
		fh, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer fh.Close()

		scanner := bufio.NewScanner(fh)
		for scanner.Scan() {
			add(scanner.Text())
		}

		err = scanner.Err()
		if err != nil {
			panic(err)
		}
	}

	return ids
}

func main() {
	flag.Parse()

	if *inputPath == "" {
		panic("input path (-i) required")
	}
	if *outputPath == "" {
		panic("output path (-o) required")
	}

	err := mmm.Filter(*outputPath, *inputPath,
		getIds(*rowsFlag, *rowsPathFlag), *rowsInverted,
		getIds(*colsFlag, *colsPathFlag), *colsInverted)
	if err != nil {
		panic(err)
	}
}
