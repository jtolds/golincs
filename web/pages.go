// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func firstURL(r *http.Request, offset, limit, total int) string {
	if offset <= 0 {
		return ""
	}
	vals := r.URL.Query()
	vals["limit"] = []string{fmt.Sprint(limit)}
	vals["offset"] = []string{fmt.Sprint(0)}
	return "?" + vals.Encode()
}

func lastURL(r *http.Request, offset, limit, total int) string {
	if total <= 0 || offset+limit >= total {
		return ""
	}
	vals := r.URL.Query()
	vals["limit"] = []string{fmt.Sprint(limit)}
	vals["offset"] = []string{fmt.Sprint(max(total-limit, 0))}
	return "?" + vals.Encode()
}

func nextURL(r *http.Request, offset, limit, total int) string {
	if total > 0 && offset+limit >= total {
		return ""
	}
	vals := r.URL.Query()
	vals["limit"] = []string{fmt.Sprint(limit)}
	vals["offset"] = []string{fmt.Sprint(max(offset+limit, 0))}
	return "?" + vals.Encode()
}

func prevURL(r *http.Request, offset, limit, total int) string {
	if offset <= 0 {
		return ""
	}
	vals := r.URL.Query()
	vals["limit"] = []string{fmt.Sprint(limit)}
	vals["offset"] = []string{fmt.Sprint(max(offset-limit, 0))}
	return "?" + vals.Encode()
}

type pageURLs struct {
	offset, limit           int
	First, Prev, Next, Last string
}

func newPageURLs(r *http.Request, offset, limit, total int) *pageURLs {
	return &pageURLs{
		offset: offset,
		limit:  limit,
		First:  firstURL(r, offset, limit, total),
		Prev:   prevURL(r, offset, limit, total),
		Next:   nextURL(r, offset, limit, total),
		Last:   lastURL(r, offset, limit, total),
	}
}

func (p *pageURLs) Render() template.HTML {
	rv := fmt.Sprintf("Page %d", p.offset/p.limit+1)
	if p.First != "" {
		rv += `| <a href="` + p.First + `">First</a>`
	}
	if p.Prev != "" {
		rv += `| <a href="` + p.Prev + `">Prev</a>`
	}
	if p.Next != "" {
		rv += `| <a href="` + p.Next + `">Next</a>`
	}
	if p.Last != "" {
		rv += `| <a href="` + p.Last + `">Last</a>`
	}
	return template.HTML(rv)
}
