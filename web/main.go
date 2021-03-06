// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/jtolds/golincs/web/dbs"
	"github.com/jtolds/golincs/web/dbs/lincs_gse92742_v0"
	"gopkg.in/webhelp.v1/whfatal"
	"gopkg.in/webhelp.v1/whlog"
	"gopkg.in/webhelp.v1/whmux"
	"gopkg.in/webhelp.v1/whredir"
	"gopkg.in/webhelp.v1/whroute"
)

var (
	listenAddr = flag.String("addr", ":8080", "address to listen on")

	sampleId  = whmux.NewStringArg()
	geneSigId = whmux.NewStringArg()
	genesetId = whmux.NewStringArg()
)

func main() {
	flag.Parse()

	lincs_92742, err := lincs_gse92742_v0.New()
	if err != nil {
		panic(err)
	}

	datasetMux := whmux.Dir{"": whredir.RedirectHandler("/")}
	datasets := []dbs.Dataset{lincs_92742}
	for id, dataset := range datasets {
		endpoints := NewEndpoints(struct {
			dbs.Dataset
			Id int
		}{Dataset: dataset, Id: id})

		datasetMux[fmt.Sprint(id)] = whmux.Dir{
			"": whmux.Exact(http.HandlerFunc(endpoints.Dataset)),

			"sample": sampleId.Shift(
				whmux.Exact(http.HandlerFunc(endpoints.Sample))),
			"genesig": geneSigId.Shift(
				whmux.Exact(http.HandlerFunc(endpoints.GeneSig))),
			"geneset": genesetId.Shift(
				whmux.Exact(http.HandlerFunc(endpoints.Geneset))),

			"search": whmux.Dir{
				"keyword":   whmux.Exact(http.HandlerFunc(endpoints.Keyword)),
				"signature": whmux.Exact(http.HandlerFunc(endpoints.Signature)),
			},
		}
	}

	routes := whlog.LogRequests(whlog.Default, whfatal.Catch(whmux.Dir{
		"": whmux.Exact(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				Render("datasets", map[string]interface{}{"datasets": datasets})
			})),
		"dataset": datasetMux,
	}))
	switch flag.Arg(0) {
	case "serve":
		panic(whlog.ListenAndServe(*listenAddr, routes))
	case "routes":
		whroute.PrintRoutes(os.Stdout, routes)
	default:
		fmt.Printf("Usage: %s <serve|routes>\n", os.Args[0])
	}
}
