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

func mean(vals []float32) (rv float64) {
	for _, val := range vals {
		rv += float64(val)
	}
	rv /= float64(len(vals))
	return rv
}

func scale(vals []float32, weight float32) []float32 {
	rv := make([]float32, len(vals))
	for i, val := range vals {
		rv[i] = val * weight
	}
	return rv
}

func (s *geneset) Query() ([]dbs.Dimension, error) {
	genes := make([]dbs.Gene, 0, len(s.genes))
	for _, gene := range s.genes {
		genes = append(genes, dbs.Gene{Name: gene, Weight: 1})
	}
	return s.ds.CombineGenes(genes)
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

func (ds *Dataset) CombineGenes(genes []dbs.Gene) ([]dbs.Dimension, error) {
	var vectors [][]float32
	for _, gene := range genes {
		if id, exists := ds.geneSigsByName[gene.Name]; exists {
			if vals, found := ds.genesigs.RowById(id); found {
				vectors = append(vectors, scale(vals, float32(gene.Weight)))
			}
		}
	}
	rv := make([]dbs.Dimension, 0, ds.genesigs.Cols())
	for i := 0; i < ds.genesigs.Cols(); i++ {
		var vals []float32
		for _, vec := range vectors {
			vals = append(vals, vec[i])
		}
		rv = append(rv, dbs.Dimension{
			Name:  ds.dimensionMap[i],
			Value: mean(vals)})
	}
	return rv, nil
}
