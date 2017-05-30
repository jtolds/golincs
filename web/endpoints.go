// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package main

import (
	"net/http"
	"strings"

	"github.com/jtolds/golincs/web/dbs"
	"golang.org/x/net/context"
	"gopkg.in/webhelp.v1/whcompat"
	"gopkg.in/webhelp.v1/wherr"
	"gopkg.in/webhelp.v1/whfatal"
	"gopkg.in/webhelp.v1/whparse"
)

type Endpoints struct {
	data dbs.Dataset
}

func NewEndpoints(data dbs.Dataset) *Endpoints {
	return &Endpoints{data: data}
}

func (a *Endpoints) sample(ctx context.Context) dbs.Sample {
	sample, err := a.data.Get(sampleId.Get(ctx))
	if err != nil {
		whfatal.Error(err)
	}
	return sample
}

func (a *Endpoints) Dataset(w http.ResponseWriter, r *http.Request) {
	offset := whparse.OptInt(r.FormValue("offset"), 0)
	limit := whparse.OptInt(r.FormValue("limit"), 30)
	samples, err := a.data.List(offset, limit)
	if err != nil {
		whfatal.Error(err)
	}
	Render("dataset", map[string]interface{}{
		"dataset":   a.data,
		"page_urls": newPageURLs(r, offset, limit, a.data.Samples()),
		"samples":   samples,
	})
}

func (a *Endpoints) Sample(w http.ResponseWriter, r *http.Request) {
	sample := a.sample(whcompat.Context(r))
	Render("sample", map[string]interface{}{
		"dataset": a.data,
		"sample":  sample,
	})
}

func (a *Endpoints) Similar(w http.ResponseWriter, r *http.Request) {
	sample := a.sample(whcompat.Context(r))
	dims, err := sample.Data()
	if err != nil {
		whfatal.Error(err)
	}
	offset := whparse.OptInt(r.FormValue("offset"), 0)
	limit := whparse.OptInt(r.FormValue("limit"), 30)
	nearest, err := a.data.Nearest(dims, nil, nil, offset, limit)
	if err != nil {
		whfatal.Error(err)
	}
	Render("similar", map[string]interface{}{
		"dataset":   a.data,
		"page_urls": newPageURLs(r, offset, limit, a.data.Samples()),
		"sample":    sample,
		"nearest":   nearest,
	})
}

func (a *Endpoints) Enriched(w http.ResponseWriter, r *http.Request) {
	sample := a.sample(whcompat.Context(r))
	dims, err := sample.Data()
	if err != nil {
		whfatal.Error(err)
	}
	offset := whparse.OptInt(r.FormValue("offset"), 0)
	limit := whparse.OptInt(r.FormValue("limit"), 30)
	enriched, err := a.data.Enriched(dims, offset, limit)
	if err != nil {
		whfatal.Error(err)
	}
	Render("enriched", map[string]interface{}{
		"dataset":   a.data,
		"page_urls": newPageURLs(r, offset, limit, a.data.Genesets()),
		"sample":    sample,
		"enriched":  enriched,
	})
}

func (a *Endpoints) parseDims(r *http.Request) ([]dbs.Dimension, error) {
	up_regulated_strings := strings.Fields(r.FormValue("up-regulated"))
	down_regulated_strings := strings.Fields(r.FormValue("down-regulated"))
	total := len(up_regulated_strings) + len(down_regulated_strings)
	if total == 0 {
		return nil, wherr.BadRequest.New("no dimensions provided")
	}
	seen := make(map[string]bool, total)
	dims := make([]dbs.Dimension, 0, total)
	for _, name := range up_regulated_strings {
		if seen[name] {
			return nil, wherr.BadRequest.New(
				"dimension %#v provided twice", name)
		}
		seen[name] = true
		dims = append(dims, dbs.Dimension{Name: name, Value: a.data.DimMax()})
	}
	for _, name := range down_regulated_strings {
		if seen[name] {
			return nil, wherr.BadRequest.New(
				"dimension %#v provided twice", name)
		}
		seen[name] = true
		dims = append(dims, dbs.Dimension{Name: name, Value: -a.data.DimMax()})
	}
	return dims, nil
}

func (a *Endpoints) Nearest(w http.ResponseWriter, r *http.Request) {
	var filters []dbs.SampleFilter
	for _, filter_string := range strings.Fields(r.FormValue("filters")) {
		parts := strings.Split(filter_string, "=")
		if len(parts) != 2 {
			whfatal.Error(wherr.BadRequest.New("bad filter"))
		}
		filters = append(filters, func(s dbs.Sample) bool {
			val, ok := s.Tags()[parts[0]]
			return !ok || strings.Contains(
				strings.ToLower(val), strings.ToLower(parts[1]))
		})
	}

	dims, err := a.parseDims(r)
	if err != nil {
		whfatal.Error(err)
	}

	offset := whparse.OptInt(r.FormValue("offset"), 0)
	limit := whparse.OptInt(r.FormValue("limit"), 30)
	nearest, err := a.data.Nearest(dims, dbs.CombineSampleFilters(filters...),
		nil, offset, limit)
	if err != nil {
		whfatal.Error(err)
	}

	Render("results", map[string]interface{}{
		"dataset":   a.data,
		"page_urls": newPageURLs(r, offset, limit, a.data.Samples()),
		"results":   nearest,
	})
}

func (a *Endpoints) EnrichedSearch(w http.ResponseWriter, r *http.Request) {
	dims, err := a.parseDims(r)
	if err != nil {
		whfatal.Error(err)
	}
	offset := whparse.OptInt(r.FormValue("offset"), 0)
	limit := whparse.OptInt(r.FormValue("limit"), 30)
	enriched, err := a.data.Enriched(dims, offset, limit)
	if err != nil {
		whfatal.Error(err)
	}
	Render("enriched_search", map[string]interface{}{
		"dataset":   a.data,
		"page_urls": newPageURLs(r, offset, limit, a.data.Genesets()),
		"enriched":  enriched,
	})
}

func (a *Endpoints) Search(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		whfatal.Error(wherr.BadRequest.New("no name provided"))
	}
	offset := whparse.OptInt(r.FormValue("offset"), 0)
	limit := whparse.OptInt(r.FormValue("limit"), 30)
	results, err := a.data.Search(name, nil, offset, limit)
	if err != nil {
		whfatal.Error(err)
	}

	Render("results", map[string]interface{}{
		"dataset":   a.data,
		"page_urls": newPageURLs(r, offset, limit, -1),
		"results":   results,
	})
}
