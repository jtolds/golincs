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

type GeneSig interface {
	Name() string
	Data() ([]Dimension, error)
}

type ScoredGeneSig interface {
	GeneSig
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
	DimMax() float64
	SampleTagNames() []string

	Samples() int
	GeneSigs() int
	Genesets() int

	ListSamples(offset, limit int) ([]Sample, error)
	ListGeneSigs(offset, limit int) ([]GeneSig, error)
	ListGenesets(offset, limit int) ([]Geneset, error)

	GetSample(sampleId string) (Sample, error)
	GetGeneSig(geneSigId string) (GeneSig, error)
	GetGeneset(genesetId string) (Geneset, error)

	NearestSamples(dims []Dimension, f1 SampleFilter, f2 ScoreFilter,
		offset, limit int) ([]ScoredSample, error)
	NearestGeneSigs(dims []Dimension, f2 ScoreFilter, offset, limit int) (
		[]ScoredGeneSig, error)
	NearestGenesets(dims []Dimension, f ScoreFilter, offset, limit int) (
		[]ScoredGeneset, error)

	SearchSamples(keyword string, filter SampleFilter, offset, limit int) (
		[]Sample, error)
	SearchGeneSigs(keyword string, offset, limit int) ([]GeneSig, error)
	SearchGenesets(keyword string, offset, limit int) ([]Geneset, error)
}
