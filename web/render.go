// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package main

import (
	"net/http"

	"github.com/jtolds/golincs/web/internal/tmpl"
	"gopkg.in/webhelp.v1/whfatal"
)

type PageCtx struct {
	Page map[string]interface{}
}

func Render(templateName string, page map[string]interface{}) {
	whfatal.Fatal(func(w http.ResponseWriter, r *http.Request) {
		tmpl.T.Render(w, r, templateName, PageCtx{Page: page})
	})
}
