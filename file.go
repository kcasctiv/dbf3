package dbf3

import (
	"errors"
	"io"
	"math"
	"os"
)

type file struct {
	hdr *header  // Header
	fld []*field // Fields
	dt  []byte   // Rows
}

func (f *file) Header() Header {
	return f.hdr
}

func (f *file) Fields() []Field {
	fields := make([]Field, len(f.fld))

	for idx := range f.fld {
		fields[idx] = f.fld[idx]
	}

	return fields
}

func (f *file) Row(idx int) (Row, error) {
	if idx < 0 || f.hdr.rows <= uint32(idx) {
		return nil, errors.New("Out of range")
	}

	offset := int(f.hdr.rlen) * idx
	return &row{fld: f.fld, dt: f.dt[offset : offset+int(f.hdr.rlen)]}, nil
}

func (f *file) NewRow() (int, error) {
	if f.hdr.rows == math.MaxUint32 {
		return 0, errors.New("Cannot add more rows")
	}
	r := make([]byte, f.hdr.rlen)
	r[0] = valid
	f.dt = append(f.dt, r...)
	f.hdr.rows++
	return int(f.hdr.rows) - 1, nil
}

func (f *file) DelRow(idx int) error {
	if idx < 0 || idx >= int(f.hdr.rows) {
		return errors.New("Out of range")
	}

	dtidx := idx * int(f.hdr.rlen)
	if f.dt[dtidx] == deleted {
		return errors.New("Already deleted")
	}
	f.dt[dtidx] = deleted
	return nil
}

func (f *file) AddField(name string, typ FieldType, length, dec byte) error {
	//TODO:
	return errors.New("Not implemeted")
}

func (f *file) DelField(field string) error {
	//TODO:
	return errors.New("Not implemeted")
}

func (f *file) Value(row int, field string) (string, error) {
	// TODO:
	return "", errors.New("Not implemented")
}

func (f *file) Set(row int, field, value string) error {
	//TODO:
	return errors.New("Not implemeted")
}

func (f *file) Save(w io.Writer) error {
	//TODO:
	return errors.New("Not implemeted")
}

func (f *file) SaveFile(fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	return f.Save(file)
}
