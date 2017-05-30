// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package lincs_gse92742_v0

import (
	"strings"

	"github.com/jtolds/golincs/mmm"
	"github.com/jtolds/golincs/web/dbs"
)

func (ds *Dataset) search(h *mmm.Handle, keyword string,
	filter dbs.SampleFilter, offset, limit int, tags bool) (rv []scoredSample,
	err error) {
	rows, err := ds.tx.Query(
		"SELECT s.id FROM signatures s, sig sig WHERE s.sig_id = sig.sig_id AND "+
			"instr(lower(sig.pert_iname), ?)", strings.ToLower(keyword))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skipped := 0
	for rows.Next() {
		var mmm_id mmm.Ident
		err = rows.Scan(&mmm_id)
		if err != nil {
			return nil, err
		}

		s, found, err := ds.load(h, mmm_id, tags)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}

		if filter != nil && !filter(s) {
			continue
		}

		if skipped < offset {
			skipped++
			continue
		}

		rv = append(rv, scoredSample{idx: -1, score: 0, Sample: s})
		if len(rv) >= limit {
			return rv, nil
		}
	}

	return rv, nil
}

func (ds *Dataset) SearchGeneSigs(keyword string, offset, limit int) (
	[]dbs.ScoredGeneSig, error) {
	rv, err := ds.search(ds.genesigs, keyword, nil, offset, limit, false)
	return scoredSamplesToScoredGeneSigs(rv), err
}

func (ds *Dataset) SearchSamples(keyword string, filter dbs.SampleFilter,
	offset, limit int) ([]dbs.ScoredSample, error) {
	rv, err := ds.search(ds.samples, keyword, filter, offset, limit, true)
	return scoredSamplesToScoredSamples(rv), err
}
