package dbf3

import (
	"errors"
	"io"
	"math"
	"os"
	"strings"
	"time"
)

const (
	eof   byte = 0x1a
	hterm byte = 0x0d
)

type file struct {
	// general data
	header header   // Header
	fields []*field // Fields
	data   []byte   // Rows + EOF

	fieldsIdx map[string]int
	converter textConverter
}

func (f *file) Rows() int          { return int(f.header.rows) }
func (f *file) HLen() int          { return int(f.header.hlen) }
func (f *file) RLen() int          { return int(f.header.rlen) }
func (f *file) Lang() LangID       { return LangID(f.header.lang) }
func (f *file) Changed() time.Time { return f.header.changedTime() }

func (f *file) SetLang(lang LangID) {
	f.header.lang = byte(lang)
	f.converter = newCharmapsTextConverter(lang)
}

func (f *file) Fields() []Field {
	fields := make([]Field, len(f.fields))

	for idx := range f.fields {
		fields[idx] = f.fields[idx]
	}

	return fields
}

func (f *file) Row(idx int) (Row, error) {
	if idx < 0 || f.header.rows <= uint32(idx) {
		return nil, errors.New("out of range")
	}

	return &row{f, idx}, nil
}

func (f *file) NewRow() (int, error) {
	if f.header.rows == math.MaxUint32 {
		return 0, errors.New("cannot add more rows")
	}
	r := make([]byte, f.header.rlen)
	r[0] = valid
	if cap(f.data)-len(f.data) < len(r) {
		dt := make([]byte, len(f.data)+len(r))
		copy(dt, f.data[:len(f.data)-1])
		copy(dt[len(f.data)-1:], r)
		dt[len(dt)-1] = eof
		f.data = dt
	} else {
		f.data = append(f.data[:len(f.data)-1], r...)
		f.data = append(f.data, eof)
	}
	f.header.rows++
	f.header.updateChanged()
	return int(f.header.rows) - 1, nil
}

func (f *file) DelRow(idx int) error {
	if idx < 0 || idx >= int(f.header.rows) {
		return errors.New("out of range")
	}

	dtidx := idx * int(f.header.rlen)
	if f.data[dtidx] == deleted {
		return errors.New("already deleted")
	}
	f.data[dtidx] = deleted
	f.header.updateChanged()
	return nil
}

func (f *file) Deleted(idx int) (bool, error) {
	if idx < 0 || idx >= int(f.header.rows) {
		return false, errors.New("out of range")
	}

	dtidx := idx * int(f.header.rlen)
	return f.data[dtidx] == deleted, nil
}

func (f *file) AddField(name string, typ FieldType, length, dec byte) error {
	//TODO:
	return errors.New("Not implemeted")
}

func (f *file) DelField(field string) error {
	fldIdx, ok := f.fieldsIdx[field]
	if !ok {
		return errors.New("field not found")
	}

	fld := f.fields[fldIdx]
	buf := make([]byte, len(f.data)-fld.Len()*f.Rows())
	var bufOffset int
	var rowOffset int
	for i := 0; i < f.Rows(); i++ {
		copy(buf[bufOffset:], f.data[rowOffset:rowOffset+fld.offset])
		copy(
			buf[bufOffset+fld.offset:],
			f.data[rowOffset+fld.offset+fld.Len():rowOffset+f.RLen()],
		)
		bufOffset += f.RLen() - fld.Len()
		rowOffset += f.RLen()
	}
	buf[len(buf)-1] = eof
	delete(f.fieldsIdx, field)
	copy(f.fields[fldIdx:], f.fields[fldIdx+1:])
	f.fields = f.fields[:len(f.fields)-1]
	f.header.hlen -= 32
	f.header.rlen -= uint16(fld.Len())
	f.data = buf
	return nil
}

func (f *file) Get(row int, field string) (string, error) {
	if row < 0 || row >= f.Rows() {
		return "", errors.New("out of range")
	}

	fldIdx, ok := f.fieldsIdx[field]
	if !ok {
		return "", errors.New("field not found")
	}

	fld := f.fields[fldIdx]
	offset := row*f.RLen() + fld.offset
	val := strings.TrimSpace(string(f.data[offset : offset+fld.Len()]))
	return f.converter.Decode(val)
}

func (f *file) Set(row int, field, value string) error {
	if row < 0 || row >= f.Rows() {
		return errors.New("out of range")
	}

	fldIdx, ok := f.fieldsIdx[field]
	if !ok {
		return errors.New("field not found")
	}

	fld := f.fields[fldIdx]

	cval, err := f.converter.Encode(value)
	if err != nil {
		return err
	}
	if len(cval) > fld.Len() {
		return errors.New("value larger than the field length")
	}

	//TODO: types check

	offset := row*f.RLen() + fld.offset
	copy(f.data[offset:], []byte(cval))
	return nil
}

func (f *file) Save(w io.Writer) error {
	buf := make([]byte, f.HLen()+f.RLen()*f.Rows()+1)

	// header
	f.header.writeTo(buf)

	// fields
	for idx, fld := range f.fields {
		fld.descr.writeTo(buf[32+idx*32:])
	}

	// header block terminator
	buf[f.HLen()-1] = hterm

	// rows
	copy(buf[f.HLen():], f.data)

	if _, err := w.Write(buf); err != nil {
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
