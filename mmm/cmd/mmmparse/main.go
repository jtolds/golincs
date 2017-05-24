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
	outPath = flag.String("o", ".", "output file")
)

func main() {
	flag.Parse()
	if *outPath == "" {
		panic("output path (-o) required")
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	dimensions := strings.Fields(scanner.Text())
	if len(dimensions) != 2 {
		panic("malformed input")
	}

	rows, err := strconv.ParseInt(dimensions[0], 10, 64)
	if err != nil {
		panic(err)
	}
	cols, err := strconv.ParseInt(dimensions[1], 10, 64)
	if err != nil {
		panic(err)
	}

	outfh, err := mmm.Create(*outPath, rows, cols)
	if err != nil {
		panic(err)
	}
	defer outfh.Close()

	ids := outfh.RowIds()
	for i := range ids {
		ids[i] = uint32(i)
	}

	ids = outfh.ColIds()
	for i := range ids {
		ids[i] = uint32(i)
	}

	rowid := -1
	for scanner.Scan() {
		rowid += 1
		if int64(rowid) >= rows {
			panic("too many rows")
		}
		vals := strings.Fields(scanner.Text())
		if int64(len(vals)) != cols {
			panic("invalid length")
		}
		floats := outfh.RowByIdx(rowid)
		for colid, val := range vals {
			float, err := strconv.ParseFloat(val, 32)
			if err != nil {
				panic(err)
			}
			floats[colid] = float32(float)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	err = outfh.Close()
	if err != nil {
		panic(err)
	}
}
