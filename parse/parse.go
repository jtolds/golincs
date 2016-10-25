// Copyright (C) 2016 JT Olds
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package parse provides some helpers to read tabular gene expression values.
package parse

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gonum/matrix/mat64"
)

// Lines is a simple helper function that wraps some boilerplate around
// bufio.Scanner. It calls cb for each line read from in.
func Lines(in io.Reader, maxLineWidth int,
	cb func(lineno int, line string) error) error {

	scan := bufio.NewScanner(in)
	scan.Buffer(nil, maxLineWidth)
	lineno := 0
	for scan.Scan() {
		lineno += 1
		err := cb(lineno, scan.Text())
		if err != nil {
			return err
		}
	}
	return scan.Err()
}

// Parse reads in gene expression values and returns the list of genes read in,
// and matrices representing the control and experiment values. The input data
// is expected to be of the form
//
//   gene-name-header control-1 control-2 control-3 experiment-1 experiment-2
//   gene-name 1.0 2.0 3.0 1.0 2.0
//   gene-name-2 1.0 -2.0 -3.0 2.5 1.0
//   ...
//
// All fields are separated by whitespace. The header should match:
//
//   <gene-name-header> <control-name>+ <experiment-name>+
//
// where the control names have the string "control" in theme. The header line
// should be followed by one or more gene value lines, which start with a
// unique gene name, followed by a white-space separated list of floating point
// values, one per each control and experiment.
func Parse(in io.Reader, maxLineWidth int) (
	genes []string, controls, experiments *mat64.Dense, err error) {

	var isControl map[int]bool
	var controlValues, experimentValues []float64
	expectedControlCount := -1
	expectedExperimentCount := -1

	err = Lines(in, maxLineWidth, func(lineno int, line string) error {
		fields := strings.Fields(line)
		if len(fields) < 1 {
			return nil
		}
		gene := fields[0]
		values := fields[1:]

		if isControl == nil {
			// parse header
			isControl = make(map[int]bool)
			for i, value := range values {
				if strings.Contains(strings.ToLower(value), "control") {
					isControl[i] = true
				}
			}
			return nil
		}

		genes = append(genes, gene)

		var controlCount, experimentCount int
		for i, value_str := range values {
			value, err := strconv.ParseFloat(value_str, 64)
			if err != nil {
				return err
			}
			if isControl[i] {
				controlValues = append(controlValues, value)
				controlCount += 1
			} else {
				experimentValues = append(experimentValues, value)
				experimentCount += 1
			}
		}
		if expectedControlCount == -1 {
			expectedControlCount = controlCount
		}
		if expectedExperimentCount == -1 {
			expectedExperimentCount = experimentCount
		}
		if controlCount != expectedControlCount {
			return fmt.Errorf("invalid number of columns on line %d", lineno)
		}
		if experimentCount != expectedExperimentCount {
			return fmt.Errorf("invalid number of columns on line %d", lineno)
		}

		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return genes, mat64.NewDense(
			len(genes), expectedControlCount, controlValues),
		mat64.NewDense(
			len(genes), expectedExperimentCount, experimentValues),
		nil
}
