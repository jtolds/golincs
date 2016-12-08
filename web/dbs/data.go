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

type Filter func(s Sample) bool

func CombineFilters(filters ...Filter) Filter {
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
	DimMax() float64
	TagNames() []string

	List(ctoken string, limit int) (
		samples []Sample, ctokenout string, err error)
	Get(sampleId string) (Sample, error)
	Nearest(dims []Dimension, filter Filter, limit int) (
		[]ScoredSample, error)
	Search(name string, filter Filter, limit int) ([]ScoredSample, error)
}
