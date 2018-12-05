package dbf3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
)

type file struct {
	hdr *header  // Header
	fld []*field // Fields
	dt  []byte   // Rows + EOF
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
	if idx < 0 || f.hdr.RW <= uint32(idx) {
		return nil, errors.New("out of range")
	}

	return &row{f, idx}, nil
}

func (f *file) NewRow() (int, error) {
	if f.hdr.RW == math.MaxUint32 {
		return 0, errors.New("cannot add more rows")
	}
	r := make([]byte, f.hdr.RL)
	r[0] = valid
	f.dt = append(f.dt, r...)
	f.hdr.RW++
	return int(f.hdr.RW) - 1, nil
}

func (f *file) DelRow(idx int) error {
	if idx < 0 || idx >= int(f.hdr.RW) {
		return errors.New("out of range")
	}

	dtidx := idx * int(f.hdr.RL)
	if f.dt[dtidx] == deleted {
		return errors.New("already deleted")
	}
	f.dt[dtidx] = deleted
	return nil
}

func (f *file) Deleted(idx int) (bool, error) {
	if idx < 0 || idx >= int(f.hdr.RW) {
		return false, errors.New("out of range")
	}

	dtidx := idx * int(f.hdr.RL)
	return f.dt[dtidx] == deleted, nil
}

func (f *file) AddField(name string, typ FieldType, length, dec byte) error {
	//TODO:
	return errors.New("Not implemeted")
}

func (f *file) DelField(field string) error {
	//TODO:
	return errors.New("Not implemeted")
}

func (f *file) Get(row int, field string) (string, error) {
	if row < 0 || row >= int(f.hdr.RW) {
		return "", errors.New("out of range")
	}
	// TODO:

	return "", errors.New("Not implemented")
}

func (f *file) Set(row int, field, value string) error {
	//TODO:
	return errors.New("Not implemeted")
}

func (f *file) Save(w io.Writer) error {
	err := binary.Write(w, binary.LittleEndian, f.hdr)
	if err != nil {
		return err
	}

	for _, fld := range f.fld {
		err := binary.Write(w, binary.LittleEndian, fld)
		if err != nil {
			return err
		}
	}

	// header block terminator
	if _, err := w.Write([]byte{0x0d}); err != nil {
		return err
	}

	data := bytes.NewBuffer(f.dt)
	if _, err := data.WriteTo(w); err != nil {
		return err
	}

	return nil
}

func (f *file) SaveFile(fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	return f.Save(file)
}

func (f *file) fieldOffset(name string) (int, error) {
	offset := 1
	for _, fld := range f.fld {
		if fld.Name() != name {
			offset += fld.Len()
		}

		return offset, nil
	}

	return 0, errors.New("field not found")
}
