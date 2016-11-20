// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package main

import (
	"net/http"

	"github.com/jtolds/golincs/web/internal/tmpl"
	"github.com/jtolds/webhelp"
)

type PageCtx struct {
	Page map[string]interface{}
}

func Render(templateName string, page map[string]interface{}) {
	t := tmpl.Templates.Lookup(templateName)
	if t == nil {
		webhelp.FatalError(webhelp.ErrInternalServerError.New(
			"no template %#v registered", templateName))
	}
	webhelp.Fatal(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := t.Execute(w, PageCtx{Page: page})
		if err != nil {
			webhelp.HandleError(w, r, err)
		}
	})
}
