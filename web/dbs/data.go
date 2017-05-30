// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package dbs

import (
	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/errors/errhttp"
)

var (
	Err         = errors.NewClass("error")
	ErrNotFound = Err.NewClass("not found", errhttp.SetStatusCode(404))
)

type Dimension struct {
	Name  string
	Value float64
}

type Sample interface {
	Id() string
	Name() string
	Tags() map[string]string

	Data() ([]Dimension, error)
}

type ScoredSample interface {
	Sample
	Score() float64
}

type SampleFilter func(Sample) bool
type ScoreFilter func(float64) bool

func CombineSampleFilters(filters ...SampleFilter) SampleFilter {
	if len(filters) == 0 {
		return nil
	}
	return func(s Sample) bool {
		for _, filter := range filters {
			if !filter(s) {
				return false
			}
		}
		return true
	}
}

type Dataset interface {
	Name() string
	Dimensions() int
	Samples() int
	Genesets() int
	DimMax() float64
	TagNames() []string

	List(offset, limit int) (samples []Sample, err error)
	Get(sampleId string) (Sample, error)
	Nearest(dims []Dimension, f1 SampleFilter, f2 ScoreFilter,
		offset, limit int) ([]ScoredSample, error)
	Search(name string, filter SampleFilter, offset, limit int) (
		[]ScoredSample, error)
	Enriched(dims []Dimension, offset, limit int) ([]ScoredGeneset, error)
}

type Geneset interface {
	Name() string
	Description() string
	Genes() []string
}

type ScoredGeneset interface {
	Geneset
	Score() float64
}
