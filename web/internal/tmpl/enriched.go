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

<ul class="nav nav-tabs">
  <li role="presentation">
    <a href="/dataset/{{.Page.dataset.Id}}/sample/{{.Page.sample.Id}}">Data</a>
  </li>
  <li role="presentation">
    <a href="/dataset/{{.Page.dataset.Id}}/sample/{{.Page.sample.Id}}/similar">Similar Samples</a>
  </li>
  <li role="presentation" class="active">
    <a>Enriched Gene Sets</a>
  </li>
</ul>

<div class="panel panel-default">
  <div class="panel-body">
    {{ template "enriched_results" . }}
  </div>
</div>

{{ template "footer" . }}`)
