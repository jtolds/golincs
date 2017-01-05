// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package main

import (
	"net/http"

	"github.com/jtolds/golincs/web/internal/tmpl"
	"gopkg.in/webhelp.v1/wherr"
	"gopkg.in/webhelp.v1/whfatal"
)

type PageCtx struct {
	Page map[string]interface{}
}

func Render(templateName string, page map[string]interface{}) {
	t := tmpl.Templates.Lookup(templateName)
	if t == nil {
		whfatal.Error(wherr.InternalServerError.New(
			"no template %#v registered", templateName))
	}
	whfatal.Fatal(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := t.Execute(w, PageCtx{Page: page})
		if err != nil {
			wherr.Handle(w, r, err)
		}
	})
}
