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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jtolds/golincs/cds"
	"github.com/jtolds/golincs/parse"
)

var (
	inputPath    = flag.String("input", "", "input path")
	maxLineWidth = flag.Int("max_line_width", 10*1024*1024, "max line width")
)

func main() {
	flag.Parse()
	err := Main()
	if err != nil {
		panic(err)
	}
}

func Main() error {
	if *inputPath == "" {
		return fmt.Errorf("--input required")
	}

	fh, err := os.Open(*inputPath)
	if err != nil {
		return err
	}
	defer fh.Close()

	genes, controls, experiments, err := parse.Parse(fh, *maxLineWidth)

	vec := cds.Compute(controls, experiments, 1)
	for i, gene := range genes {
		fmt.Printf("%s: %0.9f\n", gene, vec.At(i, 0))
	}

	return nil
}
