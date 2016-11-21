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

// Package pca provides some helpers for principal component analysis.
package pca

import (
	"fmt"
	"math"

	"github.com/gonum/matrix/mat64"
	"github.com/jtolds/golincs/mat"
)

// NIPALS calculates a limited number of principal components using the
// nonlinear iterative partial least-squares algorithm (NIPALS), developed by
// H. Wold (1966).
//
// Assuming an initial score vector u which can be arbitrarily chosen from the
// variables in the input data, the corresponding loading vector is calculated
// by the input data's transpose times u and normalized to length 1. This
// approximation can be improved by calculating the input data times the
// loading vector. Until the improvement does not exceed a small threshold,
// the improved new vector is used to repeat this procedure.
// If convergence is reached, the first principal component is subtracted from
// the currently used matrix, and the resulting residual matrix is used to find
// the second PC. This procedure is repeated until the desired number of PCs
// is reached or until the residual matrix contains very small values.
//
// NIPALS will return two matrixes, one containing the scores for each
// component and one with the loadings for each component. Last, it will return
// a vector containing how much variance each principal component represents.
//
// Read more about the NIPALS algorithm (and the code this was ported from)
// at https://cran.r-project.org/web/packages/chemometrics/vignettes/chemometrics-vignette.pdf
func NIPALS(data *mat64.Dense, components int, iterations int64,
	tolerance float64) (scores mat64.Dense, loadings mat64.Dense,
	explainedVariance mat64.Vector) {

	obsCount, varCount := data.Dims()

	varMeans := mat.ColumnMeans(data)

	var Xh mat64.Dense
	Xh.Apply(func(i, j int, v float64) float64 {
		return v - varMeans.At(j, 0)
	}, data)

	T := mat64.NewDense(obsCount, components, nil)
	P := mat64.NewDense(varCount, components, nil)
	explainedVariance = *mat64.NewVector(varCount, nil)
	var XhSquared mat64.Dense
	XhSquared.MulElem(&Xh, &Xh)
	varTotal := mat64.Sum(mat.ColumnMeans(&XhSquared))
	currVar := varTotal
	nr := int64(0)

	for h := 0; h < components; h++ {
		th := Xh.ColView(0)
		ende := false
		var ph mat64.Dense

		for !ende {
			nr += 1

			ph.Mul(Xh.T(), th)

			denom := mat64.Dot(th, th)
			ph.Apply(func(i, j int, v float64) float64 { return v / denom }, &ph)

			denom = math.Sqrt(mat64.Dot(ph.ColView(0), ph.ColView(0)))
			ph.Apply(func(i, j int, v float64) float64 { return v / denom }, &ph)

			var thnew mat64.Dense
			thnew.Mul(&Xh, &ph)
			denom = mat64.Dot(ph.ColView(0), ph.ColView(0))
			thnew.Apply(func(i, j int, v float64) float64 {
				return v / denom
			}, &thnew)
			var diff mat64.Vector
			diff.SubVec(thnew.ColView(0), th)
			prec := mat64.Dot(&diff, &diff)
			th = thnew.ColView(0)

			if prec <= math.Pow(tolerance, 2) {
				ende = true
			}
			if iterations <= nr {
				ende = true
				fmt.Println("Iteration stops without convergence")
			}
		}

		var sub mat64.Dense
		sub.Mul(th, ph.T())
		Xh.Sub(&Xh, &sub)
		T.SetCol(h, th.RawVector().Data)
		P.SetCol(h, ph.ColView(0).RawVector().Data)
		oldVar := currVar
		XhSquared.MulElem(&Xh, &Xh)
		currVar = mat64.Sum(mat.ColumnMeans(&XhSquared))
		explainedVariance.SetVec(h, (oldVar-currVar)/varTotal)
		nr = 0
	}

	return *T, *P, explainedVariance
}
