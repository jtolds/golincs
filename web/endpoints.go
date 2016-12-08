// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package main

import (
	"net/http"
	"strings"

	"github.com/jtolds/golincs/web/dbs"
	"github.com/jtolds/webhelp"
	"golang.org/x/net/context"
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
		webhelp.FatalError(err)
	}
	return sample
}

func (a *Endpoints) Dataset(w http.ResponseWriter, r *http.Request) {
	limit := webhelp.OptInt(r.FormValue("limit"), 30)
	samples, ctoken, err := a.data.List(r.FormValue("ctoken"), limit)
	if err != nil {
		webhelp.FatalError(err)
	}
	Render("dataset", map[string]interface{}{
		"dataset": a.data,
		"ctoken":  ctoken,
		"limit":   limit,
		"samples": samples,
	})
}

func (a *Endpoints) Sample(w http.ResponseWriter, r *http.Request) {
	sample := a.sample(webhelp.Context(r))
	Render("sample", map[string]interface{}{
		"dataset": a.data,
		"sample":  sample,
	})
}

func (a *Endpoints) Similar(w http.ResponseWriter, r *http.Request) {
	sample := a.sample(webhelp.Context(r))
	dims, err := sample.Data()
	if err != nil {
		webhelp.FatalError(err)
	}
	nearest, err := a.data.Nearest(dims, nil,
		webhelp.OptInt(r.FormValue("limit"), 30))
	if err != nil {
		webhelp.FatalError(err)
	}
	Render("similar", map[string]interface{}{
		"dataset": a.data,
		"sample":  sample,
		"nearest": nearest,
	})
}

func (a *Endpoints) Nearest(w http.ResponseWriter, r *http.Request) {
	up_regulated_strings := strings.Fields(r.FormValue("up-regulated"))
	down_regulated_strings := strings.Fields(r.FormValue("down-regulated"))
	var filters []dbs.Filter
	for _, filter_string := range strings.Fields(r.FormValue("filters")) {
		parts := strings.Split(filter_string, "=")
		if len(parts) != 2 {
			webhelp.FatalError(webhelp.ErrBadRequest.New("bad filter"))
		}
		filters = append(filters, func(s dbs.Sample) bool {
			val, ok := s.Tags()[parts[0]]
			return !ok || strings.Contains(
				strings.ToLower(val), strings.ToLower(parts[1]))
		})
	}

	total := len(up_regulated_strings) + len(down_regulated_strings)
	if total == 0 {
		webhelp.FatalError(webhelp.ErrBadRequest.New("no dimensions provided"))
	}
	seen := make(map[string]bool, total)
	dims := make([]dbs.Dimension, 0, total)
	for _, name := range up_regulated_strings {
		if seen[name] {
			webhelp.FatalError(webhelp.ErrBadRequest.New(
				"dimension %#v provided twice", name))
		}
		seen[name] = true
		dims = append(dims, dbs.Dimension{Name: name, Value: a.data.DimMax()})
	}
	for _, name := range down_regulated_strings {
		if seen[name] {
			webhelp.FatalError(webhelp.ErrBadRequest.New(
				"dimension %#v provided twice", name))
		}
		seen[name] = true
		dims = append(dims, dbs.Dimension{Name: name, Value: -a.data.DimMax()})
	}
	nearest, err := a.data.Nearest(dims, dbs.CombineFilters(filters...),
		webhelp.OptInt(r.FormValue("limit"), 30))
	if err != nil {
		webhelp.FatalError(err)
	}

	Render("results", map[string]interface{}{
		"dataset": a.data,
		"results": nearest,
	})
}

func (a *Endpoints) Search(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		webhelp.FatalError(webhelp.ErrBadRequest.New("no name provided"))
	}
	results, err := a.data.Search(name, nil,
		webhelp.OptInt(r.FormValue("limit"), 30))
	if err != nil {
		webhelp.FatalError(err)
	}

	Render("results", map[string]interface{}{
		"dataset": a.data,
		"results": results,
	})
}
