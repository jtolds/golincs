// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package tmpl

var _ = T.MustParse(`<table class="table table-striped">
<tr>
  <th>Name</th>
  <th>Score</th>
</tr>

{{ range .Page.enriched }}
<tr>
  <td>{{.Name}}</td>
  <td>{{.Score}}</td>
</tr>
{{ end }}

</table>

{{.Page.page_urls.Render}}
`)
