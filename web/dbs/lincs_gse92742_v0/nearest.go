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

	h := make(minHeap, 0, offset+limit)
	heap.Push(&h, scoredSample{idx: -1, score: math.Inf(-1)})

	for i := 0; i < mh.Rows(); i++ {
		vals := mh.RowByIdx(i)
		score := unitCosineSimilarity(query, vals)
		if score <= h[0].score {
			continue
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

		if len(h) >= cap(h) {
			heap.Pop(&h)
		}
		heap.Push(&h, scoredSample{
			idx:    i,
			score:  score,
			Sample: s})
	}

	sort.Sort(sort.Reverse(&h))

	rv := make([]scoredSample, 0, len(h))
	found := 0
	for _, el := range h {
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
		if found < offset {
			found++
			continue
		}
		rv = append(rv, el)
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
