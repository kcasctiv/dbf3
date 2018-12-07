package dbf3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/axgle/mahonia"
)

const (
	eof   byte = 0x1a
	hterm byte = 0x0d
)

type file struct {
	// general data
	hdr *header  // Header
	fld []*field // Fields
	dt  []byte   // Rows + EOF

	fldmap map[string]int
	enc    mahonia.Encoder
	dec    mahonia.Decoder
}

func (f *file) Rows() int          { return int(f.hdr.RW) }
func (f *file) RLen() int          { return int(f.hdr.RL) }
func (f *file) LangID() LangID     { return LangID(f.hdr.LD) }
func (f *file) Changed() time.Time { return f.hdr.changed() }

func (f *file) SetLangID(lang LangID) {
	charset := charsets[codepages[lang]]
	f.enc = mahonia.NewEncoder(charset)
	f.dec = mahonia.NewDecoder(charset)
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
	if cap(f.dt)-len(f.dt) < len(r) {
		dt := make([]byte, len(f.dt)+len(r))
		copy(dt, f.dt[:len(f.dt)-1])
		copy(dt[len(f.dt)-1:], r)
		dt[len(dt)-1] = eof
		f.dt = dt
	} else {
		f.dt = append(f.dt[:len(f.dt)-1], r...)
		f.dt = append(f.dt, eof)
	}
	f.hdr.RW++
	f.hdr.updateChanged()
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
	f.hdr.updateChanged()
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
	fldIdx, ok := f.fldmap[field]
	if !ok {
		return errors.New("field not found")
	}

	fld := f.fld[fldIdx]
	buf := make([]byte, len(f.dt)-fld.Len()*f.Rows())
	var bufOffset int
	var rowOffset int
	for i := 0; i < f.Rows(); i++ {
		rowOffset = i * f.RLen()
		copy(buf[bufOffset:], f.dt[rowOffset:rowOffset+fld.offset])
		copy(buf[bufOffset+fld.offset:], f.dt[rowOffset+fld.offset+fld.Len():rowOffset+f.RLen()])
		bufOffset += f.RLen() - fld.Len()
	}
	buf[len(buf)-1] = eof
	delete(f.fldmap, field)
	copy(f.fld[fldIdx:], f.fld[fldIdx+1:])
	f.fld = f.fld[:len(f.fld)-1]
	f.hdr.HL -= 32
	f.hdr.RL -= uint16(fld.Len())
	f.dt = buf
	return nil
}

func (f *file) Get(row int, field string) (string, error) {
	if row < 0 || row >= f.Rows() {
		return "", errors.New("out of range")
	}

	fldIdx, ok := f.fldmap[field]
	if !ok {
		return "", errors.New("field not found")
	}

	fld := f.fld[fldIdx]
	offset := row*f.RLen() + fld.offset
	val := strings.TrimSpace(string(f.dt[offset : offset+fld.Len()]))
	return f.dec.ConvertString(val), nil
}

func (f *file) Set(row int, field, value string) error {
	if row < 0 || row >= f.Rows() {
		return errors.New("out of range")
	}

	fldIdx, ok := f.fldmap[field]
	if !ok {
		return errors.New("field not found")
	}

	fld := f.fld[fldIdx]

	cval := f.enc.ConvertString(value)
	if len(cval) > fld.Len() {
		return errors.New("value larger than the field length")
	}

	//TODO: types check

	offset := row*f.RLen() + fld.offset
	copy(f.dt[offset:], []byte(cval))
	return nil
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
	if _, err := w.Write([]byte{hterm}); err != nil {
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
