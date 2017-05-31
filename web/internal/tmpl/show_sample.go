// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package tmpl

var _ = T.MustParse(`{{ template "header" . }}

<h1>Dataset: <a href="/dataset/{{.Page.dataset.Id}}">{{.Page.dataset.Name}}</a></h1>
<h2>Sample: {{.Page.sample.Name}}</h2>

<table class="table"><tr>
{{ range .Page.dataset.SampleTagNames }}
<th>{{.}}</th>
{{end}}
</tr><tr>
{{ $Page := .Page }}
{{ range .Page.dataset.SampleTagNames }}
<td>{{index $Page.sample.Tags .}}</td>
{{end}}
</tr></table>

<div style="text-align: right;">
  <a class="btn btn-primary"
      href="/dataset/{{.Page.dataset.Id}}/search/signature?qtype=sample&id={{$Page.sample.Id}}">
    Search by sample
  </a>
</div>

<table class="table table-striped">
<tr>
  <th>Dimension</th>
  <th>Value</th>
</tr>
{{ range .Page.sample.Data }}
<tr>
  <td>{{.Name}}</td>
  <td>{{.Value}}</td>
</tr>
{{ end }}
</table>

{{ template "footer" . }}`)
