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

type DataSet interface {
	Name() string
	Dimensions() int
	Samples() int
	DimMax() float64
	TagNames() []string

	List(ctoken string, limit int) (
		samples []Sample, ctokenout string, err error)
	Get(sampleId string) (Sample, error)
	Nearest(dims []Dimension, limit int) ([]ScoredSample, error)
}
