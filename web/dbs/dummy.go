// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package dbs

import (
	"strings"

	"gopkg.in/webhelp.v1/wherr"
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

var _ Dataset = (*dummySet)(nil)

func (d *dummySet) Name() string { return d.name }

func (d *dummySet) Dimensions() int {
	dims, _ := d.samples[0].Data()
	return len(dims)
}

func (d *dummySet) Samples() int  { return len(d.samples) }
func (d *dummySet) Genesets() int { return 0 }

func (d *dummySet) TagNames() []string {
	return []string{"Tag Default", "Specific tag"}
}

func (d *dummySet) List(offset, limit int) (samples []Sample, err error) {
	if offset != 0 {
		if offset < len(d.samples) {
			return nil, wherr.BadRequest.New("offset too small")
		}
		return nil, nil
	}
	if limit < len(d.samples) {
		return nil, wherr.BadRequest.New("limit too small")
	}
	var rv []Sample
	for i := range d.samples {
		rv = append(rv, &d.samples[i])
	}
	return rv, nil
}

func (d *dummySet) Get(sampleId string) (Sample, error) {
	for _, sample := range d.samples {
		if sample.id == sampleId {
			return &sample, nil
		}
	}
	return nil, ErrNotFound.New("%#v not found", sampleId)
}

func (d *dummySet) Nearest(dims []Dimension, f1 SampleFilter,
	f2 ScoreFilter, offset, limit int) ([]ScoredSample, error) {
	var rv []ScoredSample
	found := 0
	for i := range d.samples {
		if len(rv) >= limit {
			break
		}
		if f1 != nil && !f1(&d.samples[i]) {
			continue
		}
		if f2 != nil && !f2(d.samples[i].Score()) {
			continue
		}
		if found < offset {
			found++
			continue
		}
		rv = append(rv, &d.samples[i])
	}
	return rv, nil
}

func (d *dummySet) Search(name string, filter SampleFilter,
	offset, limit int) ([]ScoredSample, error) {
	var rv []ScoredSample
	found := 0
	for i := range d.samples {
		if len(rv) >= limit {
			break
		}
		if strings.Contains(d.samples[i].Name(), name) {
			if filter == nil || filter(&d.samples[i]) {
				if found < offset {
					found++
					continue
				}
				rv = append(rv, &d.samples[i])
			}
		}
	}
	return rv, nil
}

func (d *dummySet) DimMax() float64 { return 10 }

func (d *dummySet) Enriched(dims []Dimension, offset, limit int) (
	[]ScoredGeneset, error) {
	return nil, nil
}
