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

// Package mat provides some additional utility functions that complement
// the excellent github.com/gonum/matrix/mat64 package.
package mat

import (
	"github.com/gonum/matrix/mat64"
)

// Diag returns a vector of the matrix diagonal values, starting with (0, 0).
// If the matrix is not square, Diag pretends the matrix is square with a
// side-length of min(width, height).
func Diag(mat *mat64.Dense) *mat64.Vector {
	rows, cols := mat.Dims()
	rv := make([]float64, 0, rows)
	for i := 0; i < rows && i < cols; i++ {
		rv = append(rv, mat.At(i, i))
	}
	return mat64.NewVector(len(rv), rv)
}

// Eye returns an identity matrix of size width by width multiplied by value.
func Eye(width int, val float64) *mat64.Dense {
	data := make([]float64, width*width)
	for i := 0; i < width; i++ {
		data[i] = val
	}
	return mat64.NewDense(width, width, data)
}

// RowMeans returns a vector of all of the averages of each row.
func RowMeans(data *mat64.Dense) *mat64.Vector {
	rows, _ := data.Dims()
	rv := make([]float64, rows)
	for i := 0; i < rows; i++ {
		vec := data.RowView(i)
		rv[i] = mat64.Sum(vec) / float64(vec.Len())
	}
	return mat64.NewVector(rows, rv)
}

// ColumnMeans returns a vector of all of the averages of each column.
func ColumnMeans(data *mat64.Dense) *mat64.Vector {
	_, cols := data.Dims()
	rv := make([]float64, cols)
	for j := 0; j < cols; j++ {
		vec := data.ColView(j)
		rv[j] = mat64.Sum(vec) / float64(vec.Len())
	}
	return mat64.NewVector(cols, rv)
}
