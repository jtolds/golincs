// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package tmpl

var _ = T.MustParse(`{{ template "header" . }}

<h1>Dataset: <a href="/dataset/{{.Page.dataset.Id}}">{{.Page.dataset.Name}}</a></h1>
<h2>Gene signature: {{.Page.genesig.Name}}</h2>

<div style="text-align: right;">
  <a class="btn btn-primary"
      href="/dataset/{{.Page.dataset.Id}}/search/signature?qtype=genesig&id={{.Page.genesig.Id}}">
    Search by gene signature
  </a>
</div>

<table class="table table-striped">
<tr>
  <th>Dimension</th>
  <th>Value</th>
</tr>
{{ range .Page.genesig.Data }}
<tr>
  <td>{{.Name}}</td>
  <td>{{.Value}}</td>
</tr>
{{ end }}
</table>

{{ template "footer" . }}`)
