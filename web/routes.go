// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/jtolds/golincs/web/dbs"
	"github.com/jtolds/golincs/web/dbs/lincs_cds_v0"
	"github.com/jtolds/webhelp"
)

var (
	listenAddr = flag.String("addr", ":8080", "address to listen on")

	datasetId = webhelp.NewIntArgMux()
	sampleId  = webhelp.NewStringArgMux()
)

func main() {
	flag.Parse()

	lincs_cds, err := lincs_cds_v0.New(
		"sqlite3", "/home/jt/school/bio/tree-metadata.sqlite3",
		"/home/jt/school/bio/tree-mmap")
	if err != nil {
		panic(err)
	}

	endpoints := NewEndpoints([]dbs.DataSet{
		dbs.NewDummyDataSet("dummy dataset 1"), lincs_cds})

	routes := webhelp.LoggingHandler(webhelp.FatalHandler(webhelp.DirMux{
		"": webhelp.Exact(http.HandlerFunc(endpoints.Datasets)),

		"dataset": datasetId.ShiftOpt(
			webhelp.DirMux{
				"": webhelp.Exact(http.HandlerFunc(endpoints.Dataset)),

				"sample": sampleId.ShiftOpt(webhelp.DirMux{
					"":        webhelp.Exact(http.HandlerFunc(endpoints.Sample)),
					"similar": webhelp.Exact(http.HandlerFunc(endpoints.Similar)),
				},
					webhelp.RedirectHandlerFunc(
						func(r *http.Request) string {
							return fmt.Sprintf("/dataset/%d/", datasetId.MustGet(r.Context()))
						}),
				),

				"search": webhelp.ExactMethod("POST",
					webhelp.ExactPath(http.HandlerFunc(endpoints.Search))),
			},
			webhelp.ExactGet(webhelp.RedirectHandler("/")),
		),
	}))
	switch flag.Arg(0) {
	case "serve":
		panic(webhelp.ListenAndServe(*listenAddr, routes))
	case "routes":
		webhelp.PrintRoutes(os.Stdout, routes)
	default:
		fmt.Printf("Usage: %s <serve|routes>\n", os.Args[0])
	}
}
