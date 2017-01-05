// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package tmpl

var _ = T.MustParse(`{{ template "header" . }}

<h1>Datasets</h1>
<ul>
{{ range $index, $dataset := .Page.datasets }}
<li><a href="/dataset/{{$index}}">{{$dataset.Name}}</a></li>
{{ end }}
</ul>

{{ template "footer" . }}`)
