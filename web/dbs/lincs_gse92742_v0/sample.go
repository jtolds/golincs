// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package lincs_gse92742_v0

import (
	"fmt"
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
