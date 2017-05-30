// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package lincs_cds_gse70138_v0

import (
	"container/heap"
	"flag"
	"fmt"
	"math"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"unsafe"

	"github.com/jtolds/golincs/web/dbs"
	"github.com/jtolds/golincs/web/dbs/lincs_cds_gse70138_v0/metadb"
)

var (
	parallelism = flag.Int("parallelism", runtime.NumCPU(),
		"number of parallel queries to run")
)

const (
	samples    = 119155
	dimensions = 785
	maxNameLen = 45
	posSize    = int(unsafe.Sizeof(pos{}))
	nameSize   = int(unsafe.Sizeof(name{}))
)

type name struct {
	Name    [maxNameLen]byte
	NameLen int32
}

func (p *name) String() string { return string(p.Name[:p.NameLen]) }

type pos [dimensions]float64

func (p *pos) DistanceSquared(o *pos) (sum float64) {
	for i := 0; i < dimensions; i++ {
		delta := p[i] - o[i]
		sum += delta * delta
	}
	return sum
}

type pointDistance struct {
	Pos      *pos
	Name     *name
	Distance float64
}

type maxHeap []pointDistance

func (h *maxHeap) Len() int { return len(*h) }

func (h *maxHeap) Less(i, j int) bool {
	return (*h)[i].Distance > (*h)[j].Distance
}

func (h *maxHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *maxHeap) Push(x interface{}) {
	(*h) = append(*h, x.(pointDistance))
}

func (h *maxHeap) Pop() (i interface{}) {
	i, *h = (*h)[len(*h)-1], (*h)[:len(*h)-1]
	return i
}

func nearest(p pos, n int, points []pos, names []name,
	filter func(name string) (bool, error)) ([]pointDistance, error) {
	if n > len(points) {
		n = len(points)
	}
	h := make(maxHeap, 0, n)
	heap.Push(&h, pointDistance{Distance: math.Inf(1)})
	for i := range points {
		dist := p.DistanceSquared(&(points[i]))
		if dist < h[0].Distance {
			filtered, err := filter(names[i].String())
			if err != nil {
				return nil, err
			}
			if !filtered {
				continue
			}
			if len(h) >= cap(h) {
				heap.Pop(&h)
			}
			heap.Push(&h, pointDistance{
				Pos:      &(points[i]),
				Distance: dist,
				Name:     &(names[i])})
		}
	}
	return h, nil
}

func nearestParallel(p pos, n int, points []pos, names []name,
	filter func(name string) (bool, error)) (rv []pointDistance, err error) {
	type result struct {
		points []pointDistance
		err    error
	}
	results := make(chan result, *parallelism)

	amount_per_run := len(points) / (*parallelism)
	for i := 0; i < *parallelism; i++ {
		go func(i int) {
			points, err := nearest(p, n,
				points[amount_per_run*i:amount_per_run*(i+1)],
				names[amount_per_run*i:amount_per_run*(i+1)], filter)
			results <- result{points: points, err: err}
		}(i)
	}
	r := <-results
	if r.err != nil {
		return nil, r.err
	}
	for _, point := range r.points {
		if point.Name != nil {
			rv = append(rv, point)
		}
	}
	for i := 0; i < *parallelism-1; i++ {
		r = <-results
		if r.err != nil {
			return nil, r.err
		}
		for _, point := range r.points {
			if point.Name != nil {
				rv = append(rv, point)
			}
		}
	}

	sort.Sort(sort.Reverse((*maxHeap)(&rv)))
	if len(rv) > n {
		rv = rv[:n]
	}
	return rv, nil
}

type Dataset struct {
	db        *metadb.DB
	dimByIdx  []string
	dimByName map[string]int
	points    []pos
	names     []name
}

func New(driver, source, mmapTree string) (dbs.Dataset, error) {
	if *parallelism < 1 {
		return nil, fmt.Errorf("invalid parallelism value")
	}

	db, err := metadb.Open(driver, source, 0)
	if err != nil {
		return nil, err
	}
	count, err := db.CountSample()
	if err != nil {
		db.Close()
		return nil, err
	}
	if count != samples {
		db.Close()
		return nil, fmt.Errorf("invalid sample count")
	}

	ds := &Dataset{db: db}

	dims, err := db.GetDimensions()
	if err != nil {
		db.Close()
		return nil, err
	}
	if len(dims) != dimensions {
		db.Close()
		return nil, fmt.Errorf("invalid dimension count")
	}

	ds.dimByIdx = make([]string, len(dims))
	ds.dimByName = make(map[string]int, len(dims))
	for _, dim := range dims {
		ds.dimByIdx[dim.Idx] = dim.Name
		ds.dimByName[dim.Name] = int(dim.Idx)
	}

	fh, err := os.Open(mmapTree)
	if err != nil {
		db.Close()
		return nil, err
	}

	data_raw, err := syscall.Mmap(int(fh.Fd()), 0,
		(posSize+nameSize)*samples, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		fh.Close()
		db.Close()
		return nil, err
	}

	pos_data_raw := data_raw[:posSize*samples]
	name_data_raw := data_raw[posSize*samples:]

	header := *(*reflect.SliceHeader)(unsafe.Pointer(&pos_data_raw))
	header.Len /= posSize
	header.Cap /= posSize
	ds.points = *(*[]pos)(unsafe.Pointer(&header))

	header = *(*reflect.SliceHeader)(unsafe.Pointer(&name_data_raw))
	header.Len /= nameSize
	header.Cap /= nameSize
	ds.names = *(*[]name)(unsafe.Pointer(&header))

	return ds, nil
}

func (d *Dataset) Name() string    { return "LINCS Phase 2 CDS v0" }
func (d *Dataset) Dimensions() int { return dimensions }
func (d *Dataset) Samples() int    { return samples }
func (d *Dataset) DimMax() float64 { return 1 }

func (d *Dataset) List(offset, limit int) (
	samples []dbs.Sample, err error) {
	if offset != 0 {
		return nil, fmt.Errorf("implementation does not support offset")
	}
	result, _, err := d.db.PagedGetSamples("", limit)
	if err != nil {
		return nil, err
	}
	samples = make([]dbs.Sample, 0, len(result))
	for _, meta := range result {
		samples = append(samples, &Sample{meta: meta, d: d})
	}
	return samples, nil
}

func (d *Dataset) Get(sampleId string) (dbs.Sample, error) {
	meta, err := d.db.GetSampleBySigId(sampleId)
	if err != nil {
		return nil, err
	}
	return &Sample{meta: meta, d: d}, nil
}

func (d *Dataset) Nearest(dims []dbs.Dimension, filter dbs.SampleFilter,
	score_filter dbs.ScoreFilter, offset, limit int) (
	rv []dbs.ScoredSample, err error) {
	var p pos
	for _, dim := range dims {
		p[d.dimByName[dim.Name]] = dim.Value
	}
	newFilter := func(name string) (bool, error) {
		if filter == nil {
			return true, nil
		}
		meta, err := d.db.GetSampleBySigId(name)
		if err != nil {
			return false, err
		}
		return filter(&Sample{meta: meta, d: d}), nil
	}
	points, err := nearestParallel(p, offset+limit, d.points, d.names, newFilter)
	if err != nil {
		return nil, err
	}
	skipped := 0
	for _, pd := range points {
		if pd.Distance == 0 {
			continue
		}
		meta, err := d.db.GetSampleBySigId(pd.Name.String())
		if err != nil {
			return nil, err
		}
		if skipped < offset {
			skipped++
			continue
		}
		rv = append(rv, &Sample{meta: meta, d: d, score: pd.Distance})
	}
	return rv, nil
}

func (d *Dataset) Search(name string, filter dbs.SampleFilter,
	offset, limit int) (rv []dbs.ScoredSample, err error) {
	// TODO: make this whole function efficient
	name = strings.ToLower(name)
	samples, err := d.db.GetSamples()
	if err != nil {
		return nil, err
	}
	skipped := 0
	for _, sample := range samples {
		if strings.Contains(strings.ToLower(sample.Batch.String), name) ||
			strings.Contains(strings.ToLower(sample.CellId.String), name) ||
			strings.Contains(strings.ToLower(sample.PertDesc.String), name) ||
			strings.Contains(strings.ToLower(sample.PertDose.String), name) ||
			strings.Contains(strings.ToLower(sample.PertDoseUnit.String), name) ||
			strings.Contains(strings.ToLower(sample.PertId.String), name) ||
			strings.Contains(strings.ToLower(sample.PertTime.String), name) ||
			strings.Contains(strings.ToLower(sample.PertTimeUnit.String), name) ||
			strings.Contains(strings.ToLower(sample.PertType.String), name) ||
			strings.Contains(strings.ToLower(sample.ReplicateCount.String), name) ||
			strings.Contains(strings.ToLower(sample.SigId), name) {
			s := &Sample{meta: sample, d: d, score: 1}
			if filter != nil && !filter(s) {
				continue
			}
			if skipped < offset {
				skipped++
				continue
			}
			rv = append(rv, s)
			if len(rv) >= limit {
				break
			}
		}
	}
	return rv, nil
}

type Sample struct {
	meta  *metadb.Sample
	d     *Dataset
	score float64
}

func (s *Sample) Data() ([]dbs.Dimension, error) {
	for i, name := range s.d.names {
		if name.String() == s.meta.SigId {
			rv := make([]dbs.Dimension, dimensions)
			for j := 0; j < dimensions; j++ {
				rv[j].Value = s.d.points[i][j]
				rv[j].Name = s.d.dimByIdx[j]
			}
			return rv, nil
		}
	}
	return nil, dbs.ErrNotFound.New("not found")
}

func (s *Sample) Id() string     { return s.meta.SigId }
func (s *Sample) Name() string   { return s.meta.PertDesc.String }
func (s *Sample) Score() float64 { return s.score }

func (s *Sample) Tags() map[string]string {
	return map[string]string{
		"batch":           s.meta.Batch.String,
		"cell_id":         s.meta.CellId.String,
		"pert_desc":       s.meta.PertDesc.String,
		"pert_dose":       s.meta.PertDose.String + " " + s.meta.PertDoseUnit.String,
		"pert_id":         s.meta.PertId.String,
		"pert_time":       s.meta.PertTime.String + " " + s.meta.PertTimeUnit.String,
		"pert_type":       s.meta.PertType.String,
		"replicate_count": s.meta.ReplicateCount.String}
}

func (d *Dataset) TagNames() []string {
	return []string{"batch", "cell_id", "pert_desc", "pert_dose",
		"pert_id", "pert_time", "pert_type", "replicate_count"}
}

func (d *Dataset) Enriched(dims []dbs.Dimension, offset, limit int) (
	[]dbs.ScoredGeneset, error) {
	return nil, nil
}

func (d *Dataset) Genesets() int { return 0 }
