// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package lincs_gse92742_v0

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
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
	driver     = flag.String("gse92742.db_driver", "sqlite3", "database driver")
	samplePath = flag.String("gse92742.samples",
		"/home/jt/school/bio/gse92742/filtered-unit.mmap",
		"path to sample data")
	genesigPath = flag.String("gse92742.gene_sigs",
		"/home/jt/school/bio/gse92742/filtered-unit-sh_and_oe-grouped.mmap",
		"path to gene signature data")
	msigdb = flag.String("gse92742.msigdb",
		"/home/jt/school/bio/msigdb.v6.0.symbols.gmt", "path to gene sets (gmt)")

	logger = spacelog.GetLogger()
)

type Dataset struct {
	db *sql.DB
	tx *sql.Tx

	samples  *mmm.Handle
	genesigs *mmm.Handle
	genesets []*geneset

	dimensionMap        []string
	dimensionMapReverse map[string]int
	geneSigsByName      map[string]mmm.Ident
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

	sample_fh, err := mmm.Open(*samplePath)
	if err != nil {
		return nil, err
	}
	ds.samples = sample_fh

	genesig_fh, err := mmm.Open(*genesigPath)
	if err != nil {
		return nil, err
	}
	ds.genesigs = genesig_fh

	if genesig_fh.Cols() != sample_fh.Cols() {
		return nil, fmt.Errorf("gene sig and sample data column mismatch")
	}
	for idx, col := range genesig_fh.ColIds() {
		if sample_fh.ColIds()[idx] != col {
			return nil, fmt.Errorf("gene sig and sample data column mismatch")
		}
	}

	ds.dimensionMap = make([]string, sample_fh.Cols())
	ds.dimensionMapReverse = make(map[string]int, sample_fh.Cols())
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

		idx, found := sample_fh.ColIdxById(id)
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
		ds.dimensionMapReverse[gene_symbol] = idx
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	ds.geneSigsByName = map[string]mmm.Ident{}
	for i := 0; i < ds.genesigs.Rows(); i++ {
		s, err := ds.byIdx(ds.genesigs, i, false)
		if err != nil {
			return nil, err
		}
		ds.geneSigsByName[s.name] = s.mmm_id
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
				if _, exists := ds.geneSigsByName[gene]; exists {
					cleaned = append(cleaned, gene)
				}
			}
			if len(cleaned) > 0 {
				ds.genesets = append(ds.genesets, &geneset{
					id:    len(ds.genesets),
					name:  parts[0],
					desc:  parts[1],
					genes: cleaned,
					ds:    ds,
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
	if ds.genesigs != nil {
		errs.Add(ds.genesigs.Close())
		ds.genesigs = nil
	}
	if ds.tx != nil {
		errs.Add(ds.tx.Rollback())
		ds.tx = nil
	}
	if ds.db != nil {
		errs.Add(ds.db.Close())
		ds.db = nil
	}
	ds.genesets = nil
	ds.dimensionMap = nil
	ds.dimensionMapReverse = nil
	ds.geneSigsByName = nil
	return errs.Finalize()
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

func (ds *Dataset) GeneSigs() int {
	return ds.genesigs.Rows()
}

func (ds *Dataset) Genesets() int { return len(ds.genesets) }

func (ds *Dataset) DimMax() float64 { return 1 }

func (ds *Dataset) SampleTagNames() []string {
	return []string{
		// sig_id skipped
		"pert_id", "pert_type", "cell_id", "pert_idose",
		"pert_itime", "is_touchstone"}
}

func (ds *Dataset) list(h *mmm.Handle, offset, limit int, tags bool) (
	rv []*sample, err error) {
	for i := offset; i < offset+limit && i < h.Rows(); i++ {
		s, err := ds.byIdx(h, i, tags)
		if err != nil {
			return nil, err
		}
		rv = append(rv, s)
	}
	return rv, nil
}

func (ds *Dataset) ListGeneSigs(offset, limit int) ([]dbs.GeneSig, error) {
	rv, err := ds.list(ds.genesigs, offset, limit, false)
	return samplesToGeneSigs(rv), err
}

func (ds *Dataset) ListSamples(offset, limit int) (samples []dbs.Sample,
	err error) {
	rv, err := ds.list(ds.samples, offset, limit, true)
	return samplesToSamples(rv), err
}

func (ds *Dataset) ListGenesets(offset, limit int) (rv []dbs.Geneset,
	err error) {
	for i := offset; i < offset+limit && i < len(ds.genesets); i++ {
		rv = append(rv, ds.genesets[i])
	}
	return rv, nil
}

func (ds *Dataset) load(h *mmm.Handle, mmm_id mmm.Ident, tags bool) (
	rv *sample, found bool, err error) {
	values, found := h.RowById(mmm_id)
	if !found {
		return nil, false, nil
	}
	var pert_iname, pert_id, pert_type, cell_id, pert_idose, pert_itime,
		is_touchstone string
	if tags {
		err = ds.tx.QueryRow("SELECT sig.pert_iname, sig.pert_id, sig.pert_type, "+
			"sig.cell_id, sig.pert_idose, sig.pert_itime, sig.is_touchstone "+
			"FROM sig sig, signatures s WHERE s.sig_id = sig.sig_id AND s.id = ?",
			mmm_id).Scan(&pert_iname, &pert_id, &pert_type, &cell_id, &pert_idose,
			&pert_itime, &is_touchstone)
	} else {
		err = ds.tx.QueryRow("SELECT sig.pert_iname "+
			"FROM sig sig, signatures s WHERE s.sig_id = sig.sig_id AND s.id = ?",
			mmm_id).Scan(&pert_iname)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, err
	}
	rv = &sample{
		mmm_id:       mmm_id,
		name:         pert_iname,
		tags:         map[string]string{},
		data:         values,
		dimensionMap: ds.dimensionMap,
	}
	if tags {
		rv.tags["pert_id"] = pert_id
		rv.tags["pert_type"] = pert_type
		rv.tags["cell_id"] = cell_id
		rv.tags["pert_idose"] = pert_idose
		rv.tags["pert_itime"] = pert_itime
		rv.tags["is_touchstone"] = is_touchstone
	}
	return rv, true, nil
}

func (ds *Dataset) byIdx(h *mmm.Handle, idx int, tags bool) (*sample, error) {
	s, found, err := ds.load(h, h.RowIdByIdx(idx), tags)
	return s, notFound(found, err)
}

func (ds *Dataset) GetGeneSig(geneSigId string) (dbs.GeneSig, error) {
	id, err := strconv.ParseUint(geneSigId, 10, 32)
	if err != nil {
		return nil, err
	}
	rv, found, err := ds.load(ds.genesigs, mmm.Ident(id), false)
	return rv, notFound(found, err)
}

func (ds *Dataset) GetSample(sampleId string) (dbs.Sample, error) {
	id, err := strconv.ParseUint(sampleId, 10, 32)
	if err != nil {
		return nil, err
	}
	rv, found, err := ds.load(ds.samples, mmm.Ident(id), true)
	return rv, notFound(found, err)
}
