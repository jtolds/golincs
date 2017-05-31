// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package lincs_gse92742_v0

import (
	"container/heap"
	"math"
	"sort"

	"github.com/jtolds/golincs/mmm"
	"github.com/jtolds/golincs/web/dbs"
)

type heapimpl interface {
	Cap() int
	heap.Interface
	Sort()
	Data() []scoredSample
}

type minHeap []scoredSample

func (h *minHeap) Len() int { return len(*h) }
func (h *minHeap) Cap() int { return cap(*h) }

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

func (h *minHeap) Sort()               { sort.Sort(sort.Reverse(h)) }
func (h minHeap) Data() []scoredSample { return h }

type maxHeap []scoredSample

func (h *maxHeap) Len() int { return len(*h) }
func (h *maxHeap) Cap() int { return cap(*h) }

func (h *maxHeap) Less(i, j int) bool {
	return (*h)[i].score > (*h)[j].score
}

func (h *maxHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *maxHeap) Push(x interface{}) {
	(*h) = append(*h, x.(scoredSample))
}

func (h *maxHeap) Pop() (i interface{}) {
	i, *h = (*h)[len(*h)-1], (*h)[:len(*h)-1]
	return i
}

func (h *maxHeap) Sort()               { sort.Sort(h) }
func (h maxHeap) Data() []scoredSample { return h }

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

func (ds *Dataset) nearest(mh *mmm.Handle, dims []dbs.Dimension,
	sample_filter dbs.SampleFilter, score_filter dbs.ScoreFilter,
	offset, limit int, tags bool) ([]scoredSample, error) {

	query := make([]float32, mh.Cols())
	for _, dim := range dims {
		if idx, found := ds.dimensionMapReverse[dim.Name]; found {
			query[idx] = float32(dim.Value)
		}
	}
	normalize(query)

	var h heapimpl
	fromend := false
	if offset+limit <= mh.Rows()-offset {
		hi := make(minHeap, 0, offset+limit)
		h = &hi
		heap.Push(h, scoredSample{idx: -1, score: math.Inf(-1)})
	} else {
		fromend = true
		hi := make(maxHeap, 0, mh.Rows()-offset)
		h = &hi
		heap.Push(h, scoredSample{idx: -1, score: math.Inf(1)})
	}

	for i := 0; i < mh.Rows(); i++ {
		vals := mh.RowByIdx(i)
		score := unitCosineSimilarity(query, vals)
		if fromend {
			if score >= h.Data()[0].score {
				continue
			}
		} else {
			if score <= h.Data()[0].score {
				continue
			}
		}
		if score_filter != nil && !score_filter(score) {
			continue
		}

		var s dbs.Sample
		if sample_filter != nil {
			var err error
			s, err = ds.byIdx(mh, i, tags)
			if err != nil {
				return nil, err
			}
			if !sample_filter(s) {
				continue
			}
		}

		if h.Len() >= h.Cap() {
			heap.Pop(h)
		}
		heap.Push(h, scoredSample{
			idx:    i,
			score:  score,
			Sample: s})
	}

	h.Sort()

	rv := make([]scoredSample, 0, limit)
	found := 0
	for _, el := range h.Data() {
		if el.idx == -1 {
			continue
		}
		if el.Sample == nil {
			s, err := ds.byIdx(mh, el.idx, tags)
			if err != nil {
				return nil, err
			}
			el.Sample = s
		}
		if !fromend && found < offset {
			found++
			continue
		}
		rv = append(rv, el)
		if len(rv) >= limit {
			break
		}
	}
	return rv, nil
}

func (ds *Dataset) NearestGeneSigs(dims []dbs.Dimension,
	score_filter dbs.ScoreFilter, offset, limit int) (
	[]dbs.ScoredGeneSig, error) {
	rv, err := ds.nearest(ds.genesigs, dims, nil, score_filter, offset, limit,
		false)
	return scoredSamplesToScoredGeneSigs(rv), err
}

func (ds *Dataset) NearestSamples(dims []dbs.Dimension,
	filter dbs.SampleFilter, score_filter dbs.ScoreFilter, offset, limit int) (
	[]dbs.ScoredSample, error) {
	rv, err := ds.nearest(ds.samples, dims, filter, score_filter, offset, limit,
		true)
	return scoredSamplesToScoredSamples(rv), err
}
