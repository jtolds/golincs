// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package lincs_gse92742_v0

import (
	"bufio"
	"container/heap"
	"database/sql"
	"flag"
	"io"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/jtolds/golincs/mmm"
	"github.com/jtolds/golincs/web/dbs"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/spacelog"
)

var (
	db = flag.String("gse92742.db_path", "/home/jt/school/bio/gse92742/db.db",
		"path to connect to the metadata db")
	driver = flag.String("gse92742.db_driver", "sqlite3", "database driver")
	data   = flag.String("gse92742.data",
		"/home/jt/school/bio/gse92742/filtered-unit-sh_and_oe-grouped.mmap",
		"path to the data")
	msigdb = flag.String("gse92742.msigdb",
		"/home/jt/school/bio/msigdb.v6.0.symbols.gmt", "gene set file (gmt)")
	enrichmentAngle = flag.Float64("gse92742.angle", 0, "")

	logger = spacelog.GetLogger()
)

type geneset struct {
	name  string
	desc  string
	genes []string
}

func (g *geneset) Name() string        { return g.name }
func (g *geneset) Description() string { return g.desc }
func (g *geneset) Genes() []string {
	return append([]string(nil), g.genes...)
}

type scoredGeneset struct {
	*geneset
	score float64
}

func (s *scoredGeneset) Score() float64                  { return s.score }
func (s *scoredGeneset) Query() ([]dbs.Dimension, error) { panic("TODO") }

type Dataset struct {
	db           *sql.DB
	tx           *sql.Tx
	samples      *mmm.Handle
	dimensionMap []string
	idxMap       map[string]int
	genesets     []*geneset
}

var _ dbs.Dataset = (*Dataset)(nil)

func New() (*Dataset, error) {
	ds := &Dataset{}
	var success bool
	defer func() {
		if !success {
			ds.Close()
		}
	}()

	db, err := sql.Open(*driver, *db)
	if err != nil {
		return nil, err
	}
	ds.db = db
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	ds.tx = tx

	fh, err := mmm.Open(*data)
	if err != nil {
		return nil, err
	}
	ds.samples = fh

	ds.dimensionMap = make([]string, fh.Cols())
	ds.idxMap = make(map[string]int, fh.Cols())
	rows, err := tx.Query("SELECT id, pr_gene_id FROM dimensions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id mmm.Ident
		var gene_id string
		err = rows.Scan(&id, &gene_id)
		if err != nil {
			return nil, err
		}

		idx, found := fh.ColIdxById(id)
		if !found {
			continue
		}

		var gene_symbol string
		err := tx.QueryRow("SELECT pr_gene_symbol FROM pr_gene WHERE "+
			"pr_gene_id = ?", gene_id).Scan(&gene_symbol)
		if err != nil {
			return nil, err
		}

		ds.dimensionMap[idx] = gene_symbol
		ds.idxMap[gene_symbol] = idx
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	okay_genes := map[string]struct{}{}
	for i := 0; i < ds.samples.Rows(); i++ {
		s, err := ds.getByIdx(i)
		if err != nil {
			return nil, err
		}
		okay_genes[s.name] = struct{}{}
	}

	gsfh, err := os.Open(*msigdb)
	if err != nil {
		return nil, err
	}
	defer gsfh.Close()

	gsfhb := bufio.NewReader(gsfh)
	possible := 0
	for {
		line, err := gsfhb.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		parts := strings.Split(strings.TrimSpace(line), "\t")
		if len(parts) > 2 {
			possible += 1
			var cleaned []string
			for _, gene := range parts[2:] {
				if _, exists := okay_genes[gene]; exists {
					cleaned = append(cleaned, gene)
				}
			}
			if len(cleaned) > 0 {
				ds.genesets = append(ds.genesets, &geneset{
					name:  parts[0],
					desc:  parts[1],
					genes: cleaned,
				})
			}
		}
		if err == io.EOF {
			break
		}
	}
	logger.Noticef("loaded %d genesets out of %d possible\n",
		len(ds.genesets), possible)

	success = true
	return ds, nil
}

func (ds *Dataset) Close() error {
	var errs errors.ErrorGroup
	if ds.samples != nil {
		errs.Add(ds.samples.Close())
		ds.samples = nil
	}
	if ds.tx != nil {
		errs.Add(ds.tx.Rollback())
		ds.tx = nil
	}
	if ds.db != nil {
		errs.Add(ds.db.Close())
		ds.db = nil
	}
	return errs.Finalize()
}

func (ds *Dataset) getValues(sample_idx int) []float32 {
	return ds.samples.RowByIdx(sample_idx)
}

func (ds *Dataset) Name() string {
	return "LINCS Phase 1 v0"
}

func (ds *Dataset) Dimensions() int {
	return ds.samples.Cols()
}

func (ds *Dataset) Samples() int {
	return ds.samples.Rows()
}

func (ds *Dataset) DimMax() float64 { return 1 }

func (ds *Dataset) SampleTagNames() []string {
	return []string{
		"pert_id", "pert_type", "cell_id", "pert_idose",
		"pert_itime", "is_touchstone"}
}

type sample struct {
	idx          int
	id           string //sig_id
	name         string //pert_iname
	tags         map[string]string
	data         []float32
	dimensionMap []string
}

func (s *sample) Id() string              { return s.id }
func (s *sample) Name() string            { return s.name }
func (s *sample) Tags() map[string]string { return s.tags }
func (s *sample) Data() ([]dbs.Dimension, error) {
	rv := make([]dbs.Dimension, 0, len(s.data))
	for idx, val := range s.data {
		rv = append(rv, dbs.Dimension{
			Name:  s.dimensionMap[idx],
			Value: float64(val)})
	}
	sort.Sort(sort.Reverse(dimensionValueSorter(rv)))
	return rv, nil
}

func (ds *Dataset) ListSamples(offset, limit int) (samples []dbs.Sample,
	err error) {
	for i := offset; i < offset+limit && i < ds.samples.Rows(); i++ {
		s, err := ds.getByIdx(i)
		if err != nil {
			return nil, err
		}
		samples = append(samples, s)
	}
	return samples, nil
}

func (ds *Dataset) ListGenesets(offset, limit int) ([]dbs.Geneset, error) {
	panic("TODO")
}

func (ds *Dataset) loadSample(sig_id string, idx int) (*sample, error) {
	var pert_iname, pert_id, pert_type, cell_id, pert_idose,
		pert_itime, is_touchstone string
	err := ds.tx.QueryRow("SELECT pert_iname, pert_id, pert_type, cell_id, "+
		"pert_idose, pert_itime, is_touchstone FROM sig WHERE sig_id = ?",
		sig_id).Scan(&pert_iname, &pert_id, &pert_type, &cell_id, &pert_idose,
		&pert_itime, &is_touchstone)
	if err != nil {
		return nil, err
	}
	return &sample{
		idx:  idx,
		id:   sig_id,
		name: pert_iname,
		tags: map[string]string{
			"pert_id":       pert_id,
			"pert_type":     pert_type,
			"cell_id":       cell_id,
			"pert_idose":    pert_idose,
			"pert_itime":    pert_itime,
			"is_touchstone": is_touchstone},
		data:         ds.getValues(idx),
		dimensionMap: ds.dimensionMap}, nil
}

func (ds *Dataset) getByIdx(idx int) (*sample, error) {
	var sig_id string
	id := ds.samples.RowIdByIdx(idx)
	err := ds.tx.QueryRow("SELECT sig_id FROM signatures WHERE id = ?", id).
		Scan(&sig_id)
	if err != nil {
		return nil, err
	}
	return ds.loadSample(sig_id, idx)
}

func (ds *Dataset) GetSample(sampleId string) (dbs.Sample, error) {
	s, _, err := ds.getById(sampleId)
	return s, err
}

func (ds *Dataset) getById(sig_id string) (s *sample, found bool, err error) {
	var id mmm.Ident
	err = ds.tx.QueryRow("SELECT id FROM signatures WHERE sig_id = ?",
		sig_id).Scan(&id)
	if err != nil {
		return nil, false, err
	}
	idx, found := ds.samples.RowIdxById(id)
	if !found {
		return nil, false, nil
	}
	s, err = ds.loadSample(sig_id, idx)
	if err != nil {
		return nil, false, err
	}
	return s, true, nil
}

type sampleScore struct {
	idx   int
	score float64
	dbs.Sample
}

func (s sampleScore) Score() float64 { return s.score }

type minHeap []sampleScore

func (h *minHeap) Len() int { return len(*h) }

func (h *minHeap) Less(i, j int) bool {
	return (*h)[i].score < (*h)[j].score
}

func (h *minHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *minHeap) Push(x interface{}) {
	(*h) = append(*h, x.(sampleScore))
}

func (h *minHeap) Pop() (i interface{}) {
	i, *h = (*h)[len(*h)-1], (*h)[:len(*h)-1]
	return i
}

func magSquared(p []float32) (sum float64) {
	for _, v := range p {
		f := float64(v)
		sum += f * f
	}
	return sum
}

func unitCosineSimilarity(p1, p2 []float32) float64 {
	var num float64
	for i := range p1 {
		num += float64(p1[i]) * float64(p2[i])
	}
	// the denominator is 1 if p1 and p2 are both unit vectors, which they are
	return num
}

func normalize(vector []float32) {
	var squared_sum float64
	for _, val := range vector {
		squared_sum += float64(val) * float64(val)
	}
	if squared_sum == 0 {
		return
	}
	mag := math.Sqrt(squared_sum)
	for i := range vector {
		vector[i] = float32(float64(vector[i]) / mag)
	}
}

func equal(p1, p2 []float32) bool {
	if len(p1) != len(p2) {
		return false
	}
	for i, v := range p1 {
		if v != p2[i] {
			return false
		}
	}
	return true
}

func (ds *Dataset) NearestSamples(dims []dbs.Dimension,
	filter dbs.SampleFilter, score_filter dbs.ScoreFilter, offset, limit int) (
	[]dbs.ScoredSample, error) {

	query := make([]float32, ds.samples.Cols())
	for _, dim := range dims {
		if idx, found := ds.idxMap[dim.Name]; found {
			query[idx] = float32(dim.Value)
		}
	}
	normalize(query)

	h := make(minHeap, 0, offset+limit)
	heap.Push(&h, sampleScore{idx: -1, score: math.Inf(-1)})
	for i := 0; i < ds.samples.Rows() || i < 1; i++ {
		vals := ds.getValues(i)
		score := unitCosineSimilarity(query, vals)
		if score <= h[0].score || equal(query, vals) {
			continue
		}
		if score_filter != nil && !score_filter(score) {
			continue
		}

		var s dbs.Sample
		if filter != nil {
			var err error
			s, err = ds.getByIdx(i)
			if err != nil {
				return nil, err
			}
			if !filter(s) {
				continue
			}
		}

		if len(h) >= cap(h) {
			heap.Pop(&h)
		}
		heap.Push(&h, sampleScore{
			idx:    i,
			score:  score,
			Sample: s})
	}

	sort.Sort(sort.Reverse(&h))

	rv := make([]dbs.ScoredSample, 0, len(h))
	found := 0
	for _, el := range h {
		if el.idx == -1 {
			continue
		}
		if el.Sample == nil {
			s, err := ds.getByIdx(el.idx)
			if err != nil {
				return nil, err
			}
			el.Sample = s
		}
		if found < offset {
			found++
			continue
		}
		rv = append(rv, el)
	}

	return rv, nil
}

func (ds *Dataset) SampleSearch(name string, filter dbs.SampleFilter,
	offset, limit int) (rv []dbs.Sample, err error) {
	rows, err := ds.tx.Query(
		"SELECT sig_id FROM sig WHERE instr(lower(sig.pert_iname), ?)",
		strings.ToLower(name))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skipped := 0
	for rows.Next() {
		var sig_id string
		err = rows.Scan(&sig_id)
		if err != nil {
			return nil, err
		}

		s, found, err := ds.getById(sig_id)
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

		rv = append(rv, sampleScore{idx: s.idx, score: 1, Sample: s})
		if len(rv) >= limit {
			return rv, nil
		}
	}

	return rv, nil
}

func (ds *Dataset) NearestGenesets(dims []dbs.Dimension, f dbs.ScoreFilter,
	offset, limit int) ([]dbs.ScoredGeneset, error) {
	angle := *enrichmentAngle / 2
	filter := func(score float64) bool {
		return score >= angle || score <= -angle
	}
	if angle == 0 {
		filter = nil
	}
	data, err := ds.NearestSamples(dims, nil, filter, 0, ds.samples.Rows())
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

func (ds *Dataset) GenesetSearch(keyword string, offset, limit int) (
	[]dbs.Geneset, error) {
	panic("TODO")
}

func (ds *Dataset) GetGeneset(genesetId string) (dbs.Geneset, error) {
	panic("TODO")
}

func (ds *Dataset) Genesets() int { return len(ds.genesets) }

type dimensionValueSorter []dbs.Dimension

func (d dimensionValueSorter) Len() int      { return len(d) }
func (d dimensionValueSorter) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d dimensionValueSorter) Less(i, j int) bool {
	return d[i].Value < d[j].Value
}

type dimensionNameSorter []dbs.Dimension

func (d dimensionNameSorter) Len() int      { return len(d) }
func (d dimensionNameSorter) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d dimensionNameSorter) Less(i, j int) bool {
	return d[i].Name < d[j].Name
}

type scoredGenesetSorter []dbs.ScoredGeneset

func (d scoredGenesetSorter) Len() int      { return len(d) }
func (d scoredGenesetSorter) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d scoredGenesetSorter) Less(i, j int) bool {
	return d[i].Score() > d[j].Score()
}
