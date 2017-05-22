// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package mmm

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"
	"syscall"
	"unsafe"
)

const (
	float32Size = int(unsafe.Sizeof(float32(0)))
	uint32Size  = int(unsafe.Sizeof(uint32(0)))
	magicString = "FMJT"

	maxInt    = int((^uint(0)) >> 1)
	maxUint32 = ^uint32(0)
)

func uint32Slice(data []byte, offset, uint32count int) (rv []uint32,
	nextOffset int) {
	nextOffset = offset + uint32count*uint32Size
	r := data[offset:nextOffset]
	h := *(*reflect.SliceHeader)(unsafe.Pointer(&r))
	h.Len /= uint32Size
	h.Cap = h.Len
	return *(*[]uint32)(unsafe.Pointer(&h)), nextOffset
}

func float32Slice(data []byte, offset, float32count int) (rv []float32,
	nextOffset int) {
	nextOffset = offset + float32count*float32Size
	r := data[offset:nextOffset]
	h := *(*reflect.SliceHeader)(unsafe.Pointer(&r))
	h.Len /= float32Size
	h.Cap = h.Len
	return *(*[]float32)(unsafe.Pointer(&h)), nextOffset
}

type Handle struct {
	fh   *os.File
	data []byte

	rows, cols     int
	rowIds, colIds []uint32
	floats         []float32

	rowIdxOnce, colIdxOnce sync.Once
	rowIdToIdx, colIdToIdx map[uint32]int
}

func Create(path string, rows, cols int64) (*Handle, error) {
	if rows > int64(maxUint32) || cols > int64(maxUint32) {
		return nil, fmt.Errorf("rows or cols too large")
	}
	headerSize := (rows+cols+2)*int64(uint32Size) + int64(len(magicString))
	fullSize := headerSize + int64(float32Size)*rows*cols
	if fullSize > int64(maxInt) {
		return nil, fmt.Errorf("rows*cols too large")
	}

	fh, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	h := &Handle{fh: fh}
	defer h.Close()

	_, err = fh.Seek(fullSize-1, 0)
	if err != nil {
		return nil, err
	}
	_, err = fh.Write([]byte{0})
	if err != nil {
		return nil, err
	}
	_, err = fh.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	data, err := syscall.Mmap(int(fh.Fd()), 0, int(fullSize),
		syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	h.data = data

	copy(data[:len(magicString)], []byte(magicString))
	if string(data[:len(magicString)]) != magicString {
		return nil, fmt.Errorf("failed setting magic header")
	}

	sizeData, _ := uint32Slice(data, len(magicString), 2)
	sizeData[0] = uint32(rows)
	sizeData[1] = uint32(cols)

	err = h.Close()
	if err != nil {
		return nil, err
	}

	return Open(path)
}

func Open(path string) (h *Handle, err error) {
	fh, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	h = &Handle{fh: fh}
	defer func() {
		if err != nil {
			h.Close()
		}
	}()

	header := make([]byte, len(magicString)+2*uint32Size)
	_, err = io.ReadFull(fh, header)
	if err != nil {
		return nil, err
	}
	if string(header[:len(magicString)]) != magicString {
		return nil, fmt.Errorf("%#v not correct file format", path)
	}

	sizes, offset := uint32Slice(header, len(magicString), 2)
	h.rows, h.cols = int(sizes[0]), int(sizes[1])

	_, err = fh.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	headerSize := (int64(h.rows)+int64(h.cols)+2)*int64(uint32Size) +
		int64(len(magicString))
	fullSize := int64(float32Size)*int64(h.rows)*int64(h.cols) + headerSize

	data, err := syscall.Mmap(int(fh.Fd()), 0, int(fullSize),
		syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	h.data = data

	h.rowIds, offset = uint32Slice(data, offset, h.rows)
	h.colIds, offset = uint32Slice(data, offset, h.cols)
	h.floats, _ = float32Slice(data, offset, h.rows*h.cols)

	return h, nil
}

func (h *Handle) rowIndex() {
	h.rowIdToIdx = make(map[uint32]int, len(h.rowIds))
	for idx, id := range h.rowIds {
		h.rowIdToIdx[id] = idx
	}
}

func (h *Handle) colIndex() {
	h.colIdToIdx = make(map[uint32]int, len(h.colIds))
	for idx, id := range h.colIds {
		h.colIdToIdx[id] = idx
	}
}

func (h *Handle) Close() error {
	h.rowIds = nil
	h.colIds = nil
	h.floats = nil

	var rerr error
	if h.data != nil {
		data := h.data
		h.data = nil
		rerr = syscall.Munmap(data)
	}
	if h.fh != nil {
		fh := h.fh
		h.fh = nil
		err := fh.Close()
		if rerr == nil {
			rerr = err
		}
	}
	return rerr
}

func (h *Handle) RowByIdx(idx int) []float32 {
	if idx < 0 || idx >= h.rows {
		panic("row out of range")
	}
	return h.floats[h.cols*idx : h.cols*(idx+1)]
}

func (h *Handle) RowById(id uint32) (row []float32, found bool) {
	idx, found := h.RowIdxById(id)
	if !found {
		return nil, false
	}
	return h.RowByIdx(idx), true
}

func (h *Handle) RowIds() []uint32 {
	return h.rowIds
}

func (h *Handle) ColIds() []uint32 {
	return h.colIds
}

func (h *Handle) Rows() int {
	return h.rows
}

func (h *Handle) Cols() int {
	return h.cols
}

func (h *Handle) RowIdByIdx(idx int) (id uint32, found bool) {
	if idx < 0 || idx >= len(h.rowIds) {
		return 0, false
	}
	return h.rowIds[idx], true
}

func (h *Handle) ColIdByIdx(idx int) (id uint32, found bool) {
	if idx < 0 || idx >= len(h.rowIds) {
		return 0, false
	}
	return h.rowIds[idx], true
}

func (h *Handle) RowIdxById(id uint32) (idx int, found bool) {
	h.rowIdxOnce.Do(h.rowIndex)
	idx, found = h.rowIdToIdx[id]
	return idx, found
}

func (h *Handle) ColIdxById(id uint32) (idx int, found bool) {
	h.colIdxOnce.Do(h.colIndex)
	idx, found = h.colIdToIdx[id]
	return idx, found
}
