// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package lincs_gse92742_v0

import (
	"container/heap"
	"database/sql"
	"flag"
	"fmt"
	"math"
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
		"/home/jt/school/bio/gse92742/filtered-unit.mmap", "path to the data")
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

	var samples, dimensions int64

	err = tx.QueryRow("SELECT COUNT(*) FROM signatures").Scan(&samples)
	if err != nil {
		return nil, err
	}
	err = tx.QueryRow("SELECT COUNT(*) FROM dimensions").Scan(&dimensions)
	if err != nil {
		return nil, err
	}

	fh, err := mmm.Open(*data)
	if err != nil {
		return nil, err
	}
	ds.mmm = fh

	if int64(ds.mmm.Rows()) != samples || int64(ds.mmm.Cols()) != dimensions {
		return nil, fmt.Errorf("invalid dimensions")
	}

	ds.dimensionMap = make([]string, dimensions)
	ds.idxMap = make(map[string]int, dimensions)
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
			return nil, fmt.Errorf("missing id: %v", id)
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
		return nil, fmt.Errorf("not found")
	}
	err := ds.tx.QueryRow("SELECT sig_id FROM signatures WHERE id = ?", id).
		Scan(&sig_id)
	if err != nil {
		return nil, err
	}
	return ds.loadSample(sig_id, idx)
}

func (ds *Dataset) Get(sampleId string) (dbs.Sample, error) {
	var id uint32
	err := ds.tx.QueryRow("SELECT id FROM signatures WHERE sig_id = ?",
		sampleId).Scan(&id)
	if err != nil {
		return nil, err
	}
	idx, found := ds.mmm.RowIdxById(id)
	if !found {
		return nil, fmt.Errorf("invalid id")
	}
	return ds.loadSample(sampleId, idx)
}

type sampleDist struct {
	idx  int
	dist float32
	dbs.Sample
}

func (s sampleDist) Score() float64 { return float64(s.dist) }

type maxHeap []sampleDist

func (h *maxHeap) Len() int { return len(*h) }

func (h *maxHeap) Less(i, j int) bool {
	return (*h)[i].dist > (*h)[j].dist
}

func (h *maxHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *maxHeap) Push(x interface{}) {
	(*h) = append(*h, x.(sampleDist))
}

func (h *maxHeap) Pop() (i interface{}) {
	i, *h = (*h)[len(*h)-1], (*h)[:len(*h)-1]
	return i
}

func distSquared(p1, p2 []float32) (sum float32) {
	for i := 0; i < len(p1); i++ {
		delta := p1[i] - p2[i]
		sum += delta * delta
	}
	return sum
}

func normalize(vector []float32) {
	var squared_sum float64
	for _, val := range vector {
		squared_sum += float64(val) * float64(val)
	}
	dist := math.Sqrt(squared_sum)
	for i := range vector {
		vector[i] = float32(float64(vector[i]) / dist)
	}
}

func (ds *Dataset) Nearest(dims []dbs.Dimension, filter dbs.Filter,
	limit int) ([]dbs.ScoredSample, error) {

	query := make([]float32, ds.mmm.Cols())
	for _, dim := range dims {
		query[ds.idxMap[dim.Name]] = float32(dim.Value)
	}
	normalize(query)

	h := make(maxHeap, 0, limit)
	heap.Push(&h, sampleDist{idx: -1, dist: float32(math.Inf(1))})
	for i := 0; i < ds.mmm.Rows() || i < 1; i++ {
		dist := distSquared(query, ds.getValues(i))
		if dist >= h[0].dist || dist == 0 {
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
		heap.Push(&h, sampleDist{
			idx:    i,
			dist:   dist,
			Sample: s})
	}

	sort.Sort(sort.Reverse(&h))

	rv := make([]dbs.ScoredSample, 0, len(h))
	for _, el := range h {
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

func (ds *Dataset) Search(name string, filter dbs.Filter, limit int) (
	rv []dbs.ScoredSample, err error) {
	// TODO: this is terrible
	name = strings.ToLower(name)
	var ctoken string
	for {
		samples, ctokenout, err := ds.List(ctoken, 10)
		if err != nil {
			return nil, err
		}
		ctoken = ctokenout
		if len(samples) == 0 {
			return rv, nil
		}
		for _, sI := range samples {
			s := sI.(*sample)
			found := false
			if strings.Contains(strings.ToLower(s.id), name) ||
				strings.Contains(strings.ToLower(s.name), name) {
				found = true
			}
			if !found {
				for _, tagval := range s.tags {
					if strings.Contains(strings.ToLower(tagval), name) {
						found = true
						break
					}
				}
			}
			if !found {
				continue
			}
			if filter != nil && !filter(sI) {
				continue
			}
			rv = append(rv, sampleDist{idx: s.idx, dist: 0, Sample: sI})
			if len(rv) >= limit {
				return rv, nil
			}
		}
	}
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

func (ds *Dataset) Enriched(dims []dbs.Dimension, limit int) (
	[]dbs.GeneSet, error) {
	return nil, nil
}
