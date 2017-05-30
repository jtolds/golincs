// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

package mmm

import (
	"os"
)

func shouldKeep(selected, invert bool) bool {
	if invert {
		return selected
	}
	return !selected
}

func filterIds(dst, src []Ident, selected map[int]struct{}, invert bool) {
	dst_idx := 0
	for src_idx, id := range src {
		if _, s := selected[src_idx]; shouldKeep(s, invert) {
			dst[dst_idx] = id
			dst_idx++
		}
	}
}

func Filter(dst_path, src_path string,
	row_ids_selected []Ident, rows_inverted bool,
	col_ids_selected []Ident, cols_inverted bool) error {

	src, err := Open(src_path)
	if err != nil {
		return err
	}
	defer src.Close()

	rows_selected := make(map[int]struct{}, len(row_ids_selected))
	cols_selected := make(map[int]struct{}, len(col_ids_selected))
	for _, id := range row_ids_selected {
		if idx, found := src.RowIdxById(id); found {
			rows_selected[idx] = struct{}{}
		}
	}
	for _, id := range col_ids_selected {
		if idx, found := src.ColIdxById(id); found {
			cols_selected[idx] = struct{}{}
		}
	}

	var new_rows, new_cols int
	if rows_inverted {
		new_rows = len(rows_selected)
	} else {
		new_rows = src.Rows() - len(rows_selected)
	}
	if cols_inverted {
		new_cols = len(cols_selected)
	} else {
		new_cols = src.Cols() - len(cols_selected)
	}

	dst, err := Create(dst_path, int64(new_rows), int64(new_cols))
	if err != nil {
		return err
	}
	defer func() {
		dst.Close()
		if err != nil {
			os.Remove(dst_path)
		}
	}()

	filterIds(dst.RowIds(), src.RowIds(), rows_selected, rows_inverted)
	filterIds(dst.ColIds(), src.ColIds(), cols_selected, cols_inverted)

	dst_row_idx := 0
	for src_row_idx := 0; src_row_idx < src.Rows(); src_row_idx++ {
		_, row_selected := rows_selected[src_row_idx]
		if shouldKeep(row_selected, rows_inverted) {
			src_row := src.RowByIdx(src_row_idx)
			dst_row := dst.RowByIdx(dst_row_idx)
			dst_row_idx++

			dst_col_idx := 0
			for src_col_idx, val := range src_row {
				_, col_selected := cols_selected[src_col_idx]
				if shouldKeep(col_selected, cols_inverted) {
					dst_row[dst_col_idx] = val
					dst_col_idx++
				}
			}
		}
	}

	return dst.Close()
}
