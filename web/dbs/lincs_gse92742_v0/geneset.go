// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package lincs_gse92742_v0

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jtolds/golincs/web/dbs"
)

type geneset struct {
	id    int
	name  string
	desc  string
	genes []string
	ds    *Dataset
}

func (g *geneset) Id() string          { return fmt.Sprint(g.id) }
func (g *geneset) Name() string        { return g.name }
func (g *geneset) Description() string { return g.desc }
func (g *geneset) Genes() []string {
	return append([]string(nil), g.genes...)
}

func median(vals []float32) float64 {
	if len(vals) < 1 {
		return 0
	}
	sort.Sort(float32Sorter(vals))
	if len(vals)%2 == 1 {
		return float64(vals[len(vals)/2])
	}
	return (float64(vals[len(vals)/2-1]) + float64(vals[len(vals)/2])) / 2
}

func (s *geneset) Query() ([]dbs.Dimension, error) {
	var vectors [][]float32
	for _, gene := range s.genes {
		if id, exists := s.ds.geneSigsByName[gene]; exists {
			if vals, found := s.ds.genesigs.RowById(id); found {
				vectors = append(vectors, vals)
			}
		}
	}
	rv := make([]dbs.Dimension, 0, s.ds.genesigs.Cols())
	for i := 0; i < s.ds.genesigs.Cols(); i++ {
		var vals []float32
		for _, vec := range vectors {
			vals = append(vals, vec[i])
		}
		rv = append(rv, dbs.Dimension{
			Name:  s.ds.dimensionMap[i],
			Value: median(vals)})
	}
	return rv, nil
}

type scoredGeneset struct {
	*geneset
	score float64
}

func (s *scoredGeneset) Score() float64 { return s.score }

func (ds *Dataset) SearchGenesets(keyword string, offset, limit int) (
	rv []dbs.ScoredGeneset, err error) {
	keyword = strings.ToLower(keyword)
	skipped := 0
	for _, gs := range ds.genesets {
		if strings.Contains(strings.ToLower(gs.name), keyword) ||
			strings.Contains(strings.ToLower(gs.desc), keyword) {
			if skipped < offset {
				skipped++
				continue
			}
			rv = append(rv, &scoredGeneset{geneset: gs, score: 0})
			if len(rv) >= limit {
				break
			}
		}
	}
	return rv, nil
}

func (ds *Dataset) GetGeneset(genesetId string) (dbs.Geneset, error) {
	id, err := strconv.ParseUint(genesetId, 10, 0)
	if err != nil {
		return nil, err
	}
	return ds.genesets[id], nil
}

func (ds *Dataset) NearestGenesets(dims []dbs.Dimension, f dbs.ScoreFilter,
	offset, limit int) ([]dbs.ScoredGeneset, error) {
	data, err := ds.NearestGeneSigs(dims, nil, 0, ds.genesigs.Rows())
	if err != nil {
		return nil, err
	}

	scores := make(map[string]float64, len(data))
	for _, s := range data {
		scores[s.Name()] = s.Score()
	}

	gs_scores := make([]dbs.ScoredGeneset, 0, len(scores))
	for _, gs := range ds.genesets {
		var score float64
		for _, gene := range gs.genes {
			score += scores[gene]
		}
		score /= float64(len(gs.genes))
		gs_scores = append(gs_scores, &scoredGeneset{
			geneset: gs,
			score:   score,
		})
	}
	sort.Sort(scoredGenesetSorter(gs_scores))

	if offset >= len(gs_scores) {
		return nil, nil
	}
	gs_scores = gs_scores[offset:]

	if len(gs_scores) > limit {
		gs_scores = gs_scores[:limit]
	}

	return gs_scores, nil
}
