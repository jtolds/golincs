// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package lincs_gse92742_v0

import (
	"github.com/jtolds/golincs/web/dbs"
)

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

func notFound(found bool, err error) error {
	if err != nil {
		return err
	}
	if !found {
		return dbs.ErrNotFound.New("not found")
	}
	return nil
}

type float32Sorter []float32

func (u float32Sorter) Len() int           { return len(u) }
func (u float32Sorter) Less(i, j int) bool { return u[i] < u[j] }
func (u float32Sorter) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
