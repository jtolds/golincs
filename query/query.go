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

// Package query provides simple utilities for keeping track of a set of
// mat64.Vectors and querying them for nearness.
package query

import (
	"github.com/biogo/store/kdtree"
	"github.com/gonum/matrix/mat64"
)

// Database holds a set of mat64.Vectors
type Database struct {
	tree *kdtree.Tree
}

// NewDatabase creates a Database using the provided set of mat64.Vectors
func NewDatabase(points []*mat64.Vector) *Database {
	data := make(kdtree.Points, 0, len(points))
	for _, point := range points {
		data = append(data, kdtree.Point(point.RawVector().Data))
	}
	return &Database{tree: kdtree.New(data, false)}
}

// Nearest will return up to limit nearest mat64.Vectors to the given query
// point.
func (db *Database) Nearest(point *mat64.Vector, limit int) (
	nearest []*mat64.Vector) {
	keeper := kdtree.NewNKeeper(limit)
	db.tree.NearestSet(keeper, kdtree.Point(point.RawVector().Data))
	nearest = make([]*mat64.Vector, 0, keeper.Len())
	for keeper.Len() > 0 {
		point := keeper.Pop().(kdtree.Point)
		nearest = append(nearest, mat64.NewVector(len(point), point))
	}
	return nearest
}
