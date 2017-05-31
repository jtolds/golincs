// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package tmpl

var _ = T.MustParse(`{{ template "header" . }}

<h1>Dataset: {{.Page.dataset.Name}}</h1>
<p>Dataset is associated with {{ .Page.dataset.Dimensions }} directly measured
 genes, {{ .Page.dataset.GeneSigs }} gene signatures,
 {{ .Page.dataset.Samples }} sample signatures, and
 {{.Page.dataset.Genesets }} genesets.</p>

<h2>Search</h2>

<ul class="nav nav-tabs" role="tablist">
  <li role="presentation" class="active">
    <a href="#bysig" aria-controls="bysig" role="tab" data-toggle="tab">Signature</a>
  </li>
  <li role="presentation">
    <a href="#bytext" aria-controls="bytext" role="tab" data-toggle="tab">Text</a>
  </li>
</ul>

<div class="panel panel-default">
  <div class="panel-body">

<div class="tab-content">
  <div role="tabpanel" id="bysig" class="tab-pane fade in active">

<form method="GET" action="/dataset/{{.Page.dataset.Id}}/search/signature">
<div class="row">
<div class="col-md-6">
  <textarea name="up-regulated" class="form-control" rows="3"
      placeholder="up-regulated dimensions (whitespace separated)"></textarea>
  <br/>
</div>
<div class="col-md-6">
  <textarea name="down-regulated" class="form-control" rows="3"
      placeholder="down-regulated dimensions (whitespace separated)"></textarea>
  <br/>
</div>
</div>
<div class="row">
<div class="col-md-12 form-inline" style="text-align:right;">
  <div class="form-group">
    <label for="filters"><strong>filters: </strong></label>
    <input type="text" name="filters" class="form-control" id="filters" />
  </div>
  <div class="btn-group" data-toggle="buttons">
    <label class="btn btn-default active">
      <input type="radio" name="direct" value="yes"
          autocomplete="off" checked> Direct measurement
    </label>
    <label class="btn btn-default">
      <input type="radio" name="direct" value="no"
          autocomplete="off"> Signature combination
    </label>
  </div>
  <button type="submit" class="btn btn-primary">Search</button>
</div>
</div>
</form>

  </div>
  <div role="tabpanel" id="bytext" class="tab-pane fade in">

<form method="GET" action="/dataset/{{.Page.dataset.Id}}/search/keyword">
<div class="row">
<div class="col-md-12">
  <input type="text" name="keyword" class="form-control" />
</div>
</div>
<div class="row">
<div class="col-md-12 form-inline" style="text-align:right;">
  <button type="submit" class="btn btn-primary">Search</button>
</div>
</div>
</form>

  </div>
</div>

  </div>
</div>

<div class="row"><div class="col-md-12" style="text-align: right;">
  {{.Page.page_urls.Render}}
</div></div>

{{ $Page := .Page }}
<div class="row">

  <div class="col-md-4">
    <h2>Samples</h2>
    <ul>
    {{ range .Page.samples }}
    <li><a href="/dataset/{{$Page.dataset.Id}}/sample/{{.Id}}">{{.Name}}</a></li>
    {{ end }}
    </ul>
  </div>

  <div class="col-md-4">
    <h2>Gene signatures</h2>
    <ul>
    {{ range .Page.genesigs }}
    <li><a href="/dataset/{{$Page.dataset.Id}}/genesig/{{.Id}}">{{.Name}}</a></li>
    {{ end }}
    </ul>
  </div>

  <div class="col-md-4">
    <h2>Gene sets</h2>
    <ul>
    {{ range .Page.genesets }}
    <li><a href="/dataset/{{$Page.dataset.Id}}/geneset/{{.Id}}">{{.Name}}</a></li>
    {{ end }}
    </ul>
  </div>

</div>

{{ template "footer" . }}`)
