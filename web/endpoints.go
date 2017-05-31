// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jtolds/golincs/web/dbs"
	"gopkg.in/webhelp.v1/whcompat"
	"gopkg.in/webhelp.v1/wherr"
	"gopkg.in/webhelp.v1/whfatal"
	"gopkg.in/webhelp.v1/whparse"
)

const defaultLimit = 15

type Endpoints struct {
	data dbs.Dataset
}

func NewEndpoints(data dbs.Dataset) *Endpoints {
	return &Endpoints{data: data}
}

func (a *Endpoints) Dataset(w http.ResponseWriter, r *http.Request) {
	offset := whparse.OptInt(r.FormValue("offset"), 0)
	limit := whparse.OptInt(r.FormValue("limit"), defaultLimit)
	samples, err := a.data.ListSamples(offset, limit)
	if err != nil {
		whfatal.Error(err)
	}
	genesigs, err := a.data.ListGeneSigs(offset, limit)
	if err != nil {
		whfatal.Error(err)
	}
	genesets, err := a.data.ListGenesets(offset, limit)
	if err != nil {
		whfatal.Error(err)
	}
	Render("dataset", map[string]interface{}{
		"dataset": a.data,
		"page_urls": newPageURLs(r, offset, limit, max(
			a.data.Samples(),
			max(a.data.GeneSigs(), a.data.Genesets()))),
		"samples":  samples,
		"genesigs": genesigs,
		"genesets": genesets,
	})
}

func (a *Endpoints) Sample(w http.ResponseWriter, r *http.Request) {
	sample, err := a.data.GetSample(sampleId.Get(whcompat.Context(r)))
	if err != nil {
		whfatal.Error(err)
	}
	Render("show_sample", map[string]interface{}{
		"dataset": a.data,
		"sample":  sample,
	})
}

func (a *Endpoints) GeneSig(w http.ResponseWriter, r *http.Request) {
	genesig, err := a.data.GetGeneSig(geneSigId.Get(whcompat.Context(r)))
	if err != nil {
		whfatal.Error(err)
	}
	Render("show_genesig", map[string]interface{}{
		"dataset": a.data,
		"genesig": genesig,
	})
}

func (a *Endpoints) Geneset(w http.ResponseWriter, r *http.Request) {
	geneset, err := a.data.GetGeneset(genesetId.Get(whcompat.Context(r)))
	if err != nil {
		whfatal.Error(err)
	}
	Render("show_geneset", map[string]interface{}{
		"dataset": a.data,
		"geneset": geneset,
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

func (a *Endpoints) parseFilters(r *http.Request) dbs.SampleFilter {
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
	if len(filters) > 0 {
		return dbs.CombineSampleFilters(filters...)
	}
	return nil
}

func (a *Endpoints) Signature(w http.ResponseWriter, r *http.Request) {
	var dims []dbs.Dimension
	switch r.FormValue("qtype") {
	default:
		whfatal.Error(
			wherr.BadRequest.New("invalid qtype %q", r.FormValue("qtype")))
	case "input", "":
		var err error
		dims, err = a.parseDims(r)
		if err != nil {
			whfatal.Error(err)
		}
	case "sample":
		sample, err := a.data.GetSample(r.FormValue("id"))
		if err != nil {
			whfatal.Error(err)
		}
		dims, err = sample.Data()
		if err != nil {
			whfatal.Error(err)
		}
	case "genesig":
		genesig, err := a.data.GetGeneSig(r.FormValue("id"))
		if err != nil {
			whfatal.Error(err)
		}
		dims, err = genesig.Data()
		if err != nil {
			whfatal.Error(err)
		}
	case "geneset":
		geneset, err := a.data.GetGeneset(r.FormValue("id"))
		if err != nil {
			whfatal.Error(err)
		}
		dims, err = geneset.Query()
		if err != nil {
			whfatal.Error(err)
		}
	}

	if !whparse.OptBool(r.FormValue("direct"), true) {
		genes := make([]dbs.Gene, 0, len(dims))
		for _, dim := range dims {
			// This works since dim.Value is 1 or -1
			genes = append(genes, dbs.Gene{Name: dim.Name, Weight: dim.Value})
		}
		var err error
		dims, err = a.data.CombineGenes(genes)
		if err != nil {
			whfatal.Error(err)
		}
	}

	offset := whparse.OptInt(r.FormValue("offset"), 0)
	limit := whparse.OptInt(r.FormValue("limit"), defaultLimit)

	outmap := map[string]interface{}{
		"dataset": a.data,
		"url_for_rtype": func(rtype string) string {
			v := r.URL.Query()
			v["offset"] = []string{fmt.Sprint(0)}
			v["rtype"] = []string{rtype}
			return "?" + v.Encode()
		},
	}

	switch r.FormValue("rtype") {
	default:
		whfatal.Error(
			wherr.BadRequest.New("invalid rtype %q", r.FormValue("rtype")))
	case "samples", "":
		nearest, err := a.data.NearestSamples(dims, a.parseFilters(r), nil,
			offset, limit)
		if err != nil {
			whfatal.Error(err)
		}
		outmap["page_urls"] = newPageURLs(r, offset, limit, a.data.Samples())
		outmap["results"] = nearest
		Render("results_samples", outmap)
	case "genesigs":
		nearest, err := a.data.NearestGeneSigs(dims, nil, offset, limit)
		if err != nil {
			whfatal.Error(err)
		}
		outmap["page_urls"] = newPageURLs(r, offset, limit, a.data.GeneSigs())
		outmap["results"] = nearest
		Render("results_genesigs", outmap)
	case "genesets":
		nearest, err := a.data.NearestGenesets(dims, nil, offset, limit)
		if err != nil {
			whfatal.Error(err)
		}
		outmap["page_urls"] = newPageURLs(r, offset, limit, a.data.Genesets())
		outmap["results"] = nearest
		Render("results_genesets", outmap)
	}
}

func (a *Endpoints) Keyword(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("keyword")
	if name == "" {
		whfatal.Error(wherr.BadRequest.New("no keyword provided"))
	}

	offset := whparse.OptInt(r.FormValue("offset"), 0)
	limit := whparse.OptInt(r.FormValue("limit"), defaultLimit)
	outmap := map[string]interface{}{
		"dataset":   a.data,
		"page_urls": newPageURLs(r, offset, limit, -1),
		"url_for_rtype": func(rtype string) string {
			v := r.URL.Query()
			v["offset"] = []string{fmt.Sprint(0)}
			v["rtype"] = []string{rtype}
			return "?" + v.Encode()
		},
	}

	switch r.FormValue("rtype") {
	default:
		whfatal.Error(
			wherr.BadRequest.New("invalid rtype %q", r.FormValue("rtype")))
	case "samples", "":
		results, err := a.data.SearchSamples(name, a.parseFilters(r), offset,
			limit)
		if err != nil {
			whfatal.Error(err)
		}
		outmap["results"] = results
		Render("results_samples", outmap)
	case "genesigs":
		results, err := a.data.SearchGeneSigs(name, offset, limit)
		if err != nil {
			whfatal.Error(err)
		}
		outmap["results"] = results
		Render("results_genesigs", outmap)
	case "genesets":
		results, err := a.data.SearchGenesets(name, offset, limit)
		if err != nil {
			whfatal.Error(err)
		}
		outmap["results"] = results
		Render("results_genesets", outmap)
	}
}
