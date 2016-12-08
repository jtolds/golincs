// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package tmpl

func init() {
	register("results", `{{ template "header" . }}

<h1>Dataset: <a href="/dataset/{{.Page.dataset.Id}}">{{.Page.dataset.Name}}</a></h1>

<h2>Search results</h2>

<table class="table table-striped">
<tr>
<th>Id</th>
{{ range .Page.dataset.TagNames }}
<th>{{.}}</th>
{{end}}
<th>Score</th></tr>

{{ $page := .Page }}
{{ range .Page.results }}
<tr>
<td><a href="/dataset/{{$page.dataset.Id}}/sample/{{.Id}}">{{.Id}}</a></td>
{{ $sample := . }}
{{ range $page.dataset.TagNames }}
<td>{{index $sample.Tags .}}</td>
{{end}}
<td>{{.Score}}</td></tr>
{{ end }}
</table>

{{ template "footer" . }}`)
}
