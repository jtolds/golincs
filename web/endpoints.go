// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/jtolds/golincs/web/dbs"
	"github.com/jtolds/webhelp"
)

type Endpoints struct {
	data []dbs.DataSet
}

func NewEndpoints(data []dbs.DataSet) *Endpoints {
	return &Endpoints{data: data}
}

func (a *Endpoints) Datasets(w http.ResponseWriter, r *http.Request) {
	Render("datasets", map[string]interface{}{
		"datasets": a.data,
	})
}

func (a *Endpoints) dataset(ctx context.Context) dbs.DataSet {
	id := datasetId.MustGet(ctx)
	if id < 0 || id >= int64(len(a.data)) {
		webhelp.FatalError(webhelp.ErrNotFound.New("dataset not found"))
	}
	return struct {
		dbs.DataSet
		Id int64
	}{
		DataSet: a.data[id],
		Id:      id}
}

func (a *Endpoints) sample(ctx context.Context) dbs.Sample {
	sample, err := a.dataset(ctx).Get(sampleId.Get(ctx))
	if err != nil {
		webhelp.FatalError(err)
	}
	return sample
}

func (a *Endpoints) Dataset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dataset := a.dataset(ctx)
	limit := webhelp.OptInt(r.FormValue("limit"), 30)
	samples, ctoken, err := dataset.List(r.FormValue("ctoken"), limit)
	if err != nil {
		webhelp.FatalError(err)
	}
	Render("dataset", map[string]interface{}{
		"dataset": dataset,
		"ctoken":  ctoken,
		"limit":   limit,
		"samples": samples,
	})
}

func (a *Endpoints) Sample(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dataset := a.dataset(ctx)
	sample := a.sample(ctx)
	Render("sample", map[string]interface{}{
		"dataset": dataset,
		"sample":  sample,
	})
}

func (a *Endpoints) Similar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dataset := a.dataset(ctx)
	sample := a.sample(ctx)
	dims, err := sample.Data()
	if err != nil {
		webhelp.FatalError(err)
	}
	nearest, err := dataset.Nearest(dims,
		webhelp.OptInt(r.FormValue("limit"), 30))
	if err != nil {
		webhelp.FatalError(err)
	}
	Render("similar", map[string]interface{}{
		"dataset": dataset,
		"sample":  sample,
		"nearest": nearest,
	})
}

func (a *Endpoints) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dataset := a.dataset(ctx)
	up_regulated_strings := strings.Fields(r.FormValue("up-regulated"))
	down_regulated_strings := strings.Fields(r.FormValue("down-regulated"))
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
		dims = append(dims, dbs.Dimension{Name: name, Value: dataset.DimMax()})
	}
	for _, name := range down_regulated_strings {
		if seen[name] {
			webhelp.FatalError(webhelp.ErrBadRequest.New(
				"dimension %#v provided twice", name))
		}
		seen[name] = true
		dims = append(dims, dbs.Dimension{Name: name, Value: -dataset.DimMax()})
	}
	nearest, err := dataset.Nearest(dims,
		webhelp.OptInt(r.FormValue("limit"), 30))
	if err != nil {
		webhelp.FatalError(err)
	}

	Render("results", map[string]interface{}{
		"dataset": dataset,
		"nearest": nearest,
	})
}
