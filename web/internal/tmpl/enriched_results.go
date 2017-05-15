// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package tmpl

var _ = T.MustParse(`{{ template "header" . }}

<h1>Dataset: <a href="/dataset/{{.Page.dataset.Id}}">{{.Page.dataset.Name}}</a></h1>

<h2>Search results</h2>

<table class="table table-striped">
<tr>
<th>Id</th>
<th>Name</th>
{{ range .Page.dataset.TagNames }}
<th>{{.}}</th>
{{end}}
<th>Score</th></tr>

{{ $page := .Page }}
</table>

{{ template "footer" . }}`)
