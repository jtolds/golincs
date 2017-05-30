// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package tmpl

var _ = T.MustParse(`{{ template "header" . }}

<h1>Dataset: <a href="/dataset/{{.Page.dataset.Id}}">{{.Page.dataset.Name}}</a></h1>

<ul class="nav nav-tabs">
  <li role="presentation">
    <a href="{{call .Page.url_for_rtype "samples"}}">Samples</a>
  </li>
  <li role="presentation">
    <a href="{{call .Page.url_for_rtype "genesigs"}}">Gene signatures</a>
  </li>
  <li role="presentation" class="active">
    <a href="{{call .Page.url_for_rtype "genesets"}}">Genesets</a>
  </li>
</ul>

<div class="panel panel-default">
  <div class="panel-body">

  <div style="float: right;">
    {{.Page.page_urls.Render}}
  </div>

  <table class="table table-striped">

  <tr>
    <th>Id</th>
    <th>Name</th>
    <th>Score</th>
  </tr>

  {{ $page := .Page }}
  {{ range .Page.results }}
  <tr>
    <td><a href="/dataset/{{$page.dataset.Id}}/geneset/{{.Id}}">{{.Id}}</a></td>
    <td>{{.Name}}</td>
    <td>{{.Score}}</td>
  </tr>
  {{ end }}

  </table>

  </div>
</div>

{{ template "footer" . }}`)
