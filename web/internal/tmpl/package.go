// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package tmpl

import (
	"bytes"
	"html/template"
	"strings"

	"gopkg.in/webhelp.v1/whtmpl"
)

var T = whtmpl.NewCollection().Funcs(template.FuncMap{
	"linkify": func(val string) (template.HTML, error) {
		tmpl := `{{.}}`
		if strings.HasPrefix(strings.ToLower(val), "http") {
			tmpl = `<a href="{{.}}">{{.}}</a>`
		}
		t, err := template.New("").Parse(tmpl)
		if err != nil {
			return "", err
		}
		var buf bytes.Buffer
		err = t.Execute(&buf, val)
		if err != nil {
			return "", err
		}
		return template.HTML(buf.String()), nil
	}})
