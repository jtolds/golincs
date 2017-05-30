// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package main

import (
	"bufio"
	"flag"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/jtolds/golincs/mmm"
	"github.com/spacemonkeygo/errors"
)

var (
	groupFlag = flag.String(
		"groups", "", "path to file of newline-separated groups of "+
			"comma-separated ids")
	outputFlag = flag.String("o", "", "output path")
	inputFlag  = flag.String("i", "", "input path")
	opFlag     = flag.String(
		"op", "med", "operation to perform for grouping. can be 'med', 'mean', "+
			"'min', or 'max'")
)

func combinescal(vals ...float32) (rv float32) {
	if len(vals) == 0 {
		return 0
	}
	switch *opFlag {
	case "max":
		rv = vals[0]
		for _, v := range vals[1:] {
			if v > rv {
				rv = v
			}
		}
		return rv
	case "mean":
		var sum float64
		for _, v := range vals {
			sum += float64(v)
		}
		return float32(sum / float64(len(vals)))
	case "med":
		sort.Sort(float32Sorter(vals))
		if len(vals)%2 == 1 {
			return vals[len(vals)/2]
		}
		return (vals[len(vals)/2-1] + vals[len(vals)/2]) / 2
	case "min":
		rv = vals[0]
		for _, v := range vals[1:] {
			if v < rv {
				rv = v
			}
		}
		return rv
	default:
		panic("unknown op")
	}
}

func combinevec(dst []float32, src ...[]float32) {
	if len(src) <= 1 {
		if len(src) == 1 {
			copy(dst, src[0])
			return
		}
		for i := range dst {
			dst[i] = 0
		}
		return
	}

	vals := make([]float32, len(src))
	for i := range dst {
		for j := range src {
			vals[j] = src[j][i]
		}
		dst[i] = combinescal(vals...)
	}
}

func main() {
	flag.Parse()
	if *groupFlag == "" {
		panic("--groups required")
	}
	if *outputFlag == "" {
		panic("output path (-o) required")
	}
	if *inputFlag == "" {
		panic("input path (-i) required")
	}

	inputfh, err := mmm.Open(*inputFlag)
	if err != nil {
		panic(err)
	}
	defer inputfh.Close()

	groupfh, err := os.Open(*groupFlag)
	if err != nil {
		panic(err)
	}
	defer groupfh.Close()

	stat, err := groupfh.Stat()
	if err != nil {
		panic(err)
	}

	var groups [][]mmm.Ident
	scanner := bufio.NewScanner(groupfh)
	scanner.Buffer(nil, int(stat.Size()))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		var group []mmm.Ident
		for _, part := range strings.Split(line, ",") {
			part = strings.TrimSpace(part)
			if len(part) == 0 {
				continue
			}
			id, err := strconv.ParseUint(part, 10, 32)
			if err != nil {
				panic(err)
			}
			if _, found := inputfh.RowIdxById(mmm.Ident(id)); !found {
				continue
			}
			group = append(group, mmm.Ident(id))
		}
		if len(group) == 0 {
			continue
		}
		sort.Sort(identSorter(group))
		groups = append(groups, group)
	}

	err = scanner.Err()
	if err != nil {
		panic(err)
	}

	err = groupfh.Close()
	if err != nil {
		panic(err)
	}

	outputfh, err := mmm.Create(*outputFlag,
		int64(len(groups)), int64(inputfh.Cols()))
	if err != nil {
		panic(err)
	}
	defer outputfh.Close()

	copy(outputfh.ColIds(), inputfh.ColIds())

	combine_dst := make([]float32, inputfh.Cols())
	for group_idx, group := range groups {
		outputfh.RowIds()[group_idx] = group[0]
		var rows [][]float32
		for _, id := range group {
			row, found := inputfh.RowById(id)
			if found {
				rows = append(rows, row)
			}
		}
		combinevec(combine_dst, rows...)
		copy(outputfh.RowByIdx(group_idx), combine_dst)
	}

	var errs errors.ErrorGroup
	errs.Add(outputfh.Close())
	errs.Add(inputfh.Close())
	err = errs.Finalize()
	if err != nil {
		panic(err)
	}
}

type identSorter []mmm.Ident

func (u identSorter) Len() int           { return len(u) }
func (u identSorter) Less(i, j int) bool { return u[i] < u[j] }
func (u identSorter) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }

type float32Sorter []float32

func (u float32Sorter) Len() int           { return len(u) }
func (u float32Sorter) Less(i, j int) bool { return u[i] < u[j] }
func (u float32Sorter) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
