// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package tmpl

var _ = T.MustParse(`{{ template "header" . }}

<h1>Dataset: <a href="/dataset/{{.Page.dataset.Id}}">{{.Page.dataset.Name}}</a></h1>
<h2>Geneset: {{.Page.geneset.Name}}</h2>

<p>{{.Page.geneset.Description | linkify}}</p>

<div style="text-align: right;">
  <a class="btn btn-primary"
      href="/dataset/{{.Page.dataset.Id}}/search/signature?qtype=geneset&id={{.Page.geneset.Id}}">
    Search by geneset
  </a>
</div>

<h3>Genes</h3>

<p>
  {{ range .Page.geneset.Genes }}
    {{.}}
  {{ end }}
</p>

{{ template "footer" . }}`)
