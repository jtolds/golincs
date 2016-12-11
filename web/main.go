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

	dbDriver = flag.String("db-driver", "sqlite3", "database driver")
	dbPath   = flag.String("db", "/home/jt/school/bio/tree-metadata.sqlite3",
		"database")
	spatialDB = flag.String("spatial", "/home/jt/school/bio/tree-mmap",
		"spatial db path")

	sampleId = webhelp.NewStringArgMux()
)

func main() {
	flag.Parse()

	lincs_cds, err := lincs_cds_v0.New(*dbDriver, *dbPath, *spatialDB)
	if err != nil {
		panic(err)
	}

	datasetMux := webhelp.DirMux{"": webhelp.RedirectHandler("/")}
	datasets := []dbs.Dataset{dbs.NewDummyDataset("dummy dataset 1"), lincs_cds}
	for id, dataset := range datasets {
		endpoints := NewEndpoints(struct {
			dbs.Dataset
			Id int
		}{Dataset: dataset, Id: id})

		datasetMux[fmt.Sprint(id)] = webhelp.DirMux{
			"": webhelp.Exact(http.HandlerFunc(endpoints.Dataset)),

			"sample": sampleId.ShiftOpt(
				webhelp.DirMux{
					"":        webhelp.Exact(http.HandlerFunc(endpoints.Sample)),
					"similar": webhelp.Exact(http.HandlerFunc(endpoints.Similar)),
				},
				webhelp.RedirectHandler(fmt.Sprintf("/dataset/%d/", id)),
			),

			"search": webhelp.RequireMethod("POST",
				webhelp.ExactPath(http.HandlerFunc(endpoints.Search))),
			"nearest": webhelp.RequireMethod("POST",
				webhelp.ExactPath(http.HandlerFunc(endpoints.Nearest))),
		}
	}

	routes := webhelp.LoggingHandler(webhelp.FatalHandler(webhelp.DirMux{
		"": webhelp.Exact(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				Render("datasets", map[string]interface{}{"datasets": datasets})
			})),
		"dataset": datasetMux,
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
