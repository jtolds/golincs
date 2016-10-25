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

// Package cds provides routines related to Ma'ayan Lab's Characteristic
// Direction Signature method.
package cds

import (
	"github.com/gonum/matrix/mat64"
	"github.com/jtolds/golincs/mat"
	"github.com/jtolds/golincs/pca"
)

// Compute computes the characteristic direction vector from the given set
// of controls and experiments. Controls should have a column for every
// control and a row for every gene. Likewise experiments should have a column
// for every experiment and a row for every gene. r is a regularization term
// between 0 and 1 inclusive that smooths the covariance matrix and reduces
// potential noise in the dataset. The default value for r is 1, which means
// no regularization.
// See http://www.maayanlab.net/CD/ for more information.
func Compute(controls, experiments *mat64.Dense, r float64) *mat64.Vector {
	_, controlCount := controls.Dims()
	_, experimentCount := experiments.Dims()

	var combined mat64.Dense
	combined.Stack(controls.T(), experiments.T())

	rowCount, _ := combined.Dims()

	maxComponentsNum := 30
	if 30 > rowCount-1 {
		maxComponentsNum = rowCount - 1
	}

	scores, loadings, explainedVar := pca.NIPALS(&combined, maxComponentsNum,
		1e5, 1e-4)

	scores = *mat64.DenseCopyOf(scores.T())
	loadings = *mat64.DenseCopyOf(loadings.T())

	captured_variance := 0.0
	i := 0
	for ; i < explainedVar.Len(); i++ {
		captured_variance += explainedVar.At(i, 0)
		if captured_variance > 0.999 {
			break
		}
	}

	_, scoresCols := scores.Dims()
	_, loadingsCols := loadings.Dims()

	scores = *mat64.DenseCopyOf(scores.View(0, 0, i+1, scoresCols).T())
	loadings = *mat64.DenseCopyOf(loadings.View(0, 0, i+1, loadingsCols).T())

	var meanvec mat64.Vector
	meanvec.SubVec(mat.RowMeans(experiments), mat.RowMeans(controls))

	_, scoresCols = scores.Dims()
	ctrlScores := scores.View(0, 0, controlCount, scoresCols)
	expmScores := scores.View(controlCount, 0, experimentCount, scoresCols)

	var ctrlScoresSquared mat64.Dense
	ctrlScoresSquared.Mul(ctrlScores.T(), ctrlScores)
	var expmScoresSquared mat64.Dense
	expmScoresSquared.Mul(expmScores.T(), expmScores)
	var Dd mat64.Dense
	Dd.Add(&ctrlScoresSquared, &expmScoresSquared)
	Dd.Apply(func(i, j int, v float64) float64 {
		return v / float64(controlCount+experimentCount-2)
	}, &Dd)

	diag := mat.Diag(&Dd)
	sigma := mat64.Sum(diag) / float64(diag.Len())

	DdRows, DdCols := Dd.Dims()
	t := mat.Eye(DdRows, (1-r)*sigma)
	if DdCols != DdRows {
		panic("uh oh")
	}

	var shrunkMats mat64.Dense
	shrunkMats.Apply(func(i, j int, v float64) float64 {
		return r*v + t.At(i, j)
	}, &Dd)

	var invMat mat64.Dense
	invMat.Inverse(&shrunkMats)

	var intermediate1, intermediate2, b mat64.Dense
	intermediate1.Mul(loadings.T(), &meanvec)
	intermediate2.Mul(&invMat, &intermediate1)
	b.Mul(&loadings, &intermediate2)
	norm := mat64.Norm(&b, 2)
	b.Apply(func(i, j int, v float64) float64 {
		return v / norm
	}, &b)

	return b.ColView(0)
}
