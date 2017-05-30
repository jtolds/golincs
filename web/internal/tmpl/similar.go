// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package tmpl

var _ = T.MustParse(`{{ template "header" . }}

<h1>Dataset: <a href="/dataset/{{.Page.dataset.Id}}">{{.Page.dataset.Name}}</a></h1>
<h2>Sample: {{.Page.sample.Name}}</h2>

<table class="table"><tr>
{{ range .Page.dataset.TagNames }}
<th>{{.}}</th>
{{end}}
</tr><tr>
{{ $Page := .Page }}
{{ range .Page.dataset.TagNames }}
<td>{{index $Page.sample.Tags .}}</td>
{{end}}
</tr></table>

<ul class="nav nav-tabs">
  <li role="presentation">
    <a href="/dataset/{{.Page.dataset.Id}}/sample/{{.Page.sample.Id}}">Data</a>
  </li>
  <li role="presentation" class="active">
    <a>Similar Samples</a>
  </li>
  <li role="presentation">
    <a href="/dataset/{{.Page.dataset.Id}}/sample/{{.Page.sample.Id}}/enriched">Enriched Gene Sets</a>
  </li>
</ul>

<div class="panel panel-default">
  <div class="panel-body">

  <table class="table table-striped">

  <tr>
    <th>Id</th>
    <th>Name</th>
    {{ range .Page.dataset.TagNames }}
    <th>{{.}}</th>
    {{end}}
    <th>Score</th>
  </tr>

  {{ $page := .Page }}
  {{ range .Page.nearest }}
  <tr>
    <td><a href="/dataset/{{$page.dataset.Id}}/sample/{{.Id}}/similar">{{.Id}}</a></td>
    <td>{{.Name}}</td>
    {{ $sample := . }}
    {{ range $page.dataset.TagNames }}
      <td>{{index $sample.Tags .}}</td>
    {{end}}
    <td>{{.Score}}</td>
  </tr>
  {{ end }}

  </table>

  {{.Page.page_urls.Render}}

  </div>
</div>

{{ template "footer" . }}`)
