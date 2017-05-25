// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package lincs_gse92742_v0

import (
	"container/heap"
	"database/sql"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/jtolds/golincs/mmm"
	"github.com/jtolds/golincs/web/dbs"
	"github.com/spacemonkeygo/errors"
)

var (
	db = flag.String("gse92742.db_path", "/home/jt/school/bio/gse92742/db.db",
		"path to connect to the metadata db")
	driver = flag.String("gse92742.db_driver", "sqlite3", "database driver")
	data   = flag.String("gse92742.data",
		"/home/jt/school/bio/gse92742/filtered-unit-sh_and_oe-grouped.mmap",
		"path to the data")
	enrichmentAngle = flag.Float64("gse92742.angle", 0, "")
)

type Dataset struct {
	db           *sql.DB
	tx           *sql.Tx
	mmm          *mmm.Handle
	dimensionMap []string
	idxMap       map[string]int
}

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
	ds.mmm = fh

	ds.dimensionMap = make([]string, fh.Cols())
	ds.idxMap = make(map[string]int, fh.Cols())
	rows, err := tx.Query("SELECT id, pr_gene_id FROM dimensions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint32
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

	success = true
	return ds, nil
}

func (ds *Dataset) Close() error {
	var errs errors.ErrorGroup
	if ds.mmm != nil {
		errs.Add(ds.mmm.Close())
		ds.mmm = nil
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
	return ds.mmm.RowByIdx(sample_idx)
}

func (ds *Dataset) Name() string {
	return "LINCS Phase 1 v0"
}

func (ds *Dataset) Dimensions() int {
	return ds.mmm.Cols()
}

func (ds *Dataset) Samples() int {
	return ds.mmm.Rows()
}

func (ds *Dataset) DimMax() float64 { return 1 }

func (ds *Dataset) TagNames() []string {
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

func (ds *Dataset) List(ctoken string, limit int) (
	samples []dbs.Sample, ctokenout string, err error) {

	var offset int
	if ctoken != "" {
		offset, err = strconv.Atoi(ctoken)
		if err != nil {
			return nil, "", err
		}
	}

	for i := offset; i < offset+limit && i < ds.mmm.Rows(); i++ {
		s, err := ds.getByIdx(i)
		if err != nil {
			return nil, "", err
		}
		samples = append(samples, s)
	}

	return samples, fmt.Sprint(offset + limit), nil
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
	id, found := ds.mmm.RowIdByIdx(idx)
	if !found {
		return nil, fmt.Errorf("idx %d not found", idx)
	}
	err := ds.tx.QueryRow("SELECT sig_id FROM signatures WHERE id = ?", id).
		Scan(&sig_id)
	if err != nil {
		return nil, err
	}
	return ds.loadSample(sig_id, idx)
}

func (ds *Dataset) Get(sampleId string) (dbs.Sample, error) {
	s, _, err := ds.getById(sampleId)
	return s, err
}

func (ds *Dataset) getById(sig_id string) (s *sample, found bool, err error) {
	var id uint32
	err = ds.tx.QueryRow("SELECT id FROM signatures WHERE sig_id = ?",
		sig_id).Scan(&id)
	if err != nil {
		return nil, false, err
	}
	idx, found := ds.mmm.RowIdxById(id)
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

func (ds *Dataset) Nearest(dims []dbs.Dimension, filter dbs.SampleFilter,
	score_filter dbs.ScoreFilter, limit int) ([]dbs.ScoredSample, error) {

	query := make([]float32, ds.mmm.Cols())
	for _, dim := range dims {
		if idx, found := ds.idxMap[dim.Name]; found {
			query[idx] = float32(dim.Value)
		}
	}
	normalize(query)

	h := make(minHeap, 0, limit)
	heap.Push(&h, sampleScore{idx: -1, score: math.Inf(-1)})
	for i := 0; i < ds.mmm.Rows() || i < 1; i++ {
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
		rv = append(rv, el)
	}

	return rv, nil
}

func (ds *Dataset) Search(name string, filter dbs.SampleFilter, limit int) (
	rv []dbs.ScoredSample, err error) {
	rows, err := ds.tx.Query(
		"SELECT sig_id FROM sig WHERE instr(lower(sig.pert_iname), ?)",
		strings.ToLower(name))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

		rv = append(rv, sampleScore{idx: s.idx, score: 1, Sample: s})
		if len(rv) >= limit {
			return rv, nil
		}
	}

	return rv, nil
}

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

func (ds *Dataset) Enriched(dims []dbs.Dimension) (
	[]dbs.GeneSet, error) {
	angle := *enrichmentAngle / 2
	filter := func(score float64) bool {
		return score >= angle || score <= -angle
	}
	if angle == 0 {
		filter = nil
	}
	data, err := ds.Nearest(dims, nil, filter, ds.mmm.Rows())
	if err != nil {
		return nil, err
	}
	fh, err := os.Create("/tmp/test-data.out")
	if err != nil {
		return nil, err
	}
	_, err = fmt.Fprintf(fh, "\tBaseline\tSignature\n")
	if err != nil {
		return nil, err
	}
	for _, thing := range data {
		_, err = fmt.Fprintf(fh, "%s\t0\t%f\n", thing.Name(), thing.Score())
		if err != nil {
			return nil, err
		}
	}
	err = fh.Close()
	if err != nil {
		return nil, err
	}
	fh, err = os.Create("/tmp/test-classes.out")
	if err != nil {
		return nil, err
	}
	_, err = fmt.Fprintf(fh, "Baseline\tBaseline\nSignature\tSignature\n")
	if err != nil {
		return nil, err
	}
	err = fh.Close()
	if err != nil {
		return nil, err
	}

	return nil, nil
}
