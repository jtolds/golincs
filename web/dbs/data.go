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

type Geneset interface {
	Name() string
	Description() string
	Genes() []string

	Query() ([]Dimension, error)
}

type ScoredGeneset interface {
	Geneset
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

func CombineScoreFilters(filters ...ScoreFilter) ScoreFilter {
	if len(filters) == 0 {
		return nil
	}
	return func(s float64) bool {
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
	SampleTagNames() []string

	Genesets() int

	DimMax() float64

	ListSamples(offset, limit int) ([]Sample, error)
	ListGenesets(offset, limit int) ([]Geneset, error)

	GetSample(sampleId string) (Sample, error)
	GetGeneset(genesetId string) (Geneset, error)

	NearestSamples(dims []Dimension, f1 SampleFilter, f2 ScoreFilter,
		offset, limit int) ([]ScoredSample, error)
	NearestGenesets(dims []Dimension, f ScoreFilter, offset, limit int) (
		[]ScoredGeneset, error)

	SampleSearch(keyword string, filter SampleFilter, offset, limit int) (
		[]Sample, error)
	GenesetSearch(keyword string, offset, limit int) (
		[]Geneset, error)
}
