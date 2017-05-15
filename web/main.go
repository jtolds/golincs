// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/jtolds/golincs/web/dbs"
	"github.com/jtolds/golincs/web/dbs/lincs_cds_gse70138_v0"
	"github.com/jtolds/golincs/web/dbs/lincs_gse92742_v0"
	"gopkg.in/webhelp.v1/whfatal"
	"gopkg.in/webhelp.v1/whlog"
	"gopkg.in/webhelp.v1/whmux"
	"gopkg.in/webhelp.v1/whredir"
	"gopkg.in/webhelp.v1/whroute"
)

var (
	listenAddr = flag.String("addr", ":8080", "address to listen on")

	dbDriver    = flag.String("db-driver", "sqlite3", "database driver")
	dbPath70138 = flag.String("db70138",
		"/home/jt/school/bio/tree-metadata-70138.sqlite3", "database")
	spatialDB70138 = flag.String("spatial70138",
		"/home/jt/school/bio/tree-mmap-70138", "spatial db path")

	sampleId = whmux.NewStringArg()
)

func main() {
	flag.Parse()

	lincs_70138, err := lincs_cds_gse70138_v0.New(*dbDriver, *dbPath70138,
		*spatialDB70138)
	if err != nil {
		panic(err)
	}

	lincs_92742, err := lincs_gse92742_v0.New()
	if err != nil {
		panic(err)
	}

	datasetMux := whmux.Dir{"": whredir.RedirectHandler("/")}
	datasets := []dbs.Dataset{dbs.NewDummyDataset("dummy dataset 1"),
		lincs_70138, lincs_92742}
	for id, dataset := range datasets {
		endpoints := NewEndpoints(struct {
			dbs.Dataset
			Id int
		}{Dataset: dataset, Id: id})

		datasetMux[fmt.Sprint(id)] = whmux.Dir{
			"": whmux.Exact(http.HandlerFunc(endpoints.Dataset)),

			"sample": sampleId.ShiftOpt(
				whmux.Dir{
					"":         whmux.Exact(http.HandlerFunc(endpoints.Sample)),
					"similar":  whmux.Exact(http.HandlerFunc(endpoints.Similar)),
					"enriched": whmux.Exact(http.HandlerFunc(endpoints.Enriched)),
				},
				whredir.RedirectHandler(fmt.Sprintf("/dataset/%d/", id)),
			),

			"search": whmux.RequireMethod("POST",
				whmux.ExactPath(http.HandlerFunc(endpoints.Search))),
			"nearest": whmux.RequireMethod("POST",
				whmux.ExactPath(http.HandlerFunc(endpoints.Nearest))),
			"enriched": whmux.RequireMethod("POST",
				whmux.ExactPath(http.HandlerFunc(endpoints.EnrichedSearch))),
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
