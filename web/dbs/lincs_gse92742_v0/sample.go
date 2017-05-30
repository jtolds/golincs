// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package lincs_gse92742_v0

import (
	"fmt"
	"math"
	"sort"

	"github.com/jtolds/golincs/mmm"
	"github.com/jtolds/golincs/web/dbs"
)

type sample struct {
	mmm_id       mmm.Ident
	name         string // pert_iname
	tags         map[string]string
	data         []float32
	dimensionMap []string
}

func (s *sample) Id() string              { return fmt.Sprint(s.mmm_id) }
func (s *sample) Name() string            { return s.name }
func (s *sample) Tags() map[string]string { return s.tags }
func (s *sample) Data() ([]dbs.Dimension, error) {
	rv := make([]dbs.Dimension, 0, len(s.data))
	for idx, val := range s.data {
		rv = append(rv, dbs.Dimension{
			Name:  s.dimensionMap[idx],
			Value: float64(val)})
	}
	sort.Sort(sort.Reverse(dimensionValueSorter(rv)))
	return rv, nil
}

func samplesToGeneSigs(l []*sample) (r []dbs.GeneSig) {
	r = make([]dbs.GeneSig, len(l))
	for i, v := range l {
		r[i] = v
	}
	return r
}

func samplesToSamples(l []*sample) (r []dbs.Sample) {
	r = make([]dbs.Sample, len(l))
	for i, v := range l {
		r[i] = v
	}
	return r
}

type scoredSample struct {
	idx   int
	score float64
	dbs.Sample
}

func (s scoredSample) Score() float64 { return s.score }

type minHeap []scoredSample

func (h *minHeap) Len() int { return len(*h) }

func (h *minHeap) Less(i, j int) bool {
	return (*h)[i].score < (*h)[j].score
}

func (h *minHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *minHeap) Push(x interface{}) {
	(*h) = append(*h, x.(scoredSample))
}

func (h *minHeap) Pop() (i interface{}) {
	i, *h = (*h)[len(*h)-1], (*h)[:len(*h)-1]
	return i
}

func unitCosineSimilarity(p1, p2 []float32) float64 {
	var num float64
	for i := range p1 {
		num += float64(p1[i]) * float64(p2[i])
	}
	// the denominator is 1 if p1 and p2 are both unit vectors, which they are
	return num
}

func normalize(vector []float32) {
	var squared_sum float64
	for _, val := range vector {
		squared_sum += float64(val) * float64(val)
	}
	if squared_sum == 0 {
		return
	}
	mag := math.Sqrt(squared_sum)
	for i := range vector {
		vector[i] = float32(float64(vector[i]) / mag)
	}
}

func scoredSamplesToScoredGeneSigs(l []scoredSample) (r []dbs.ScoredGeneSig) {
	r = make([]dbs.ScoredGeneSig, len(l))
	for i, v := range l {
		r[i] = v
	}
	return r
}

func scoredSamplesToScoredSamples(l []scoredSample) (r []dbs.ScoredSample) {
	r = make([]dbs.ScoredSample, len(l))
	for i, v := range l {
		r[i] = v
	}
	return r
}
