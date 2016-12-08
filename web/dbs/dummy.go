// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package dbs

import (
	"strings"

	"github.com/jtolds/webhelp"
)

type dummySample struct {
	id     string
	name   string
	tag    string
	offset int
}

func (d *dummySample) Score() float64 { return float64(d.offset) / 10 }

func (d *dummySample) Id() string   { return d.id }
func (d *dummySample) Name() string { return d.name }
func (d *dummySample) Tags() map[string]string {
	return map[string]string{
		"Tag Default":  "a tag",
		"Specific tag": d.tag}
}

func (d *dummySample) Data() ([]Dimension, error) {
	return []Dimension{
		{Name: "dim1", Value: float64(0 + d.offset)},
		{Name: "dim2", Value: float64(1 + d.offset)}}, nil
}

type dummySet struct {
	name    string
	samples []dummySample
}

func NewDummyDataset(name string) Dataset {
	samples := []dummySample{
		{id: "1", name: "sample 1", tag: "tag 1", offset: 1},
		{id: "2", name: "sample 2", tag: "tag 2", offset: 2},
		{id: "3", name: "sample 3", tag: "tag 3", offset: 3},
	}
	return &dummySet{name: name, samples: samples}
}

func (d *dummySet) Name() string { return d.name }

func (d *dummySet) Dimensions() int {
	dims, _ := d.samples[0].Data()
	return len(dims)
}

func (d *dummySet) Samples() int { return len(d.samples) }

func (d *dummySet) TagNames() []string {
	return []string{"Tag Default", "Specific tag"}
}

func (d *dummySet) List(ctoken string, limit int) (
	samples []Sample, ctokenout string, err error) {
	if limit < len(d.samples) {
		return nil, "", webhelp.ErrBadRequest.New("limit too small")
	}
	var rv []Sample
	for i := range d.samples {
		rv = append(rv, &d.samples[i])
	}
	return rv, "", nil
}

func (d *dummySet) Get(sampleId string) (Sample, error) {
	for _, sample := range d.samples {
		if sample.id == sampleId {
			return &sample, nil
		}
	}
	return nil, ErrNotFound.New("%#v not found", sampleId)
}

func (d *dummySet) Nearest(dims []Dimension, filter Filter, limit int) (
	[]ScoredSample, error) {
	var rv []ScoredSample
	for i := range d.samples {
		if i >= limit {
			break
		}
		if filter == nil || filter(&d.samples[i]) {
			rv = append(rv, &d.samples[i])
		}
	}
	return rv, nil
}

func (d *dummySet) Search(name string, filter Filter, limit int) (
	[]ScoredSample, error) {
	var rv []ScoredSample
	for i := range d.samples {
		if i >= limit {
			break
		}
		if strings.Contains(d.samples[i].Name(), name) {
			if filter == nil || filter(&d.samples[i]) {
				rv = append(rv, &d.samples[i])
			}
		}
	}
	return rv, nil
}

func (d *dummySet) DimMax() float64 { return 10 }
