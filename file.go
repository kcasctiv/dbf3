package dbf3

import (
	"encoding/binary"
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
	f.header.updateChanged()
	f.converter = newCharmapsTextConverter(lang)
}

func (f *file) Fields() []Field {
	fields := make([]Field, len(f.fields))

	for idx := range f.fields {
		fields[idx] = f.fields[idx]
	}

	return fields
}

func (f *file) HasField(field string) bool {
	_, ok := f.fieldsIdx[field]
	return ok
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
	r := make([]byte, f.header.rlen+1)
	for idx := range r {
		r[idx] = blank
	}
	r[f.header.rlen] = eof
	f.data = append(f.data[:len(f.data)-1], r...)
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
	if !isASCII(name) {
		return errors.New("only ASCII chars allowed in field name")
	}
	name = strings.TrimSpace(name)
	if len(name) > 11 {
		return errors.New("exceeded max field name length")
	}
	if _, exists := f.fieldsIdx[name]; exists {
		return errors.New("field already exists")
	}

	switch typ {
	case Date:
		return f.addField(name, typ, 8, 0)
	case Logical:
		return f.addField(name, typ, 1, 0)
	case Numeric:
		// TODO: check length and dec
		if length-dec < 2 {
			return errors.New("decimal count must be lower at least 2 than length")
		}
		return f.addField(name, typ, length, dec)
	case Character:
		flen := binary.LittleEndian.Uint16([]byte{length, dec})
		if flen > math.MaxInt16 {
			return errors.New("exceeded max field length")
		}
		return f.addField(name, typ, length, dec)
	default:
		return errors.New("unsupported field type")
	}
}

func (f *file) addField(name string, typ FieldType, length, dec byte) error {
	dt := fieldDescr{
		len: length,
		dec: dec,
		typ: byte(typ),
	}
	copy(dt.name[:], name)
	idx := len(f.fields)
	var offset int
	if idx > 0 {
		offset = f.fields[idx-1].offset + f.fields[idx-1].Len()
	}
	fld := newField(dt, idx, offset)
	f.fields = append(f.fields, fld)
	f.fieldsIdx[fld.Name()] = idx
	f.header.hlen += 32
	f.header.rlen += uint16(fld.Len())
	f.header.updateChanged()

	if f.Rows() == 0 {
		return nil
	}

	buf := make([]byte, f.Rows()*fld.Len())
	oldLen := len(f.data)
	f.data = append(f.data, buf...)
	f.data[len(f.data)-1] = f.data[oldLen-1] // move EOF
	oldRLen := f.RLen() - fld.Len()
	for row := f.Rows() - 1; row >= 0; row-- {
		// move old row data
		copy(f.data[row*f.RLen():], f.data[row*oldRLen:(row+1)*oldRLen])
		// fill new field bytes with blank values
		for idx = row*f.RLen() + oldRLen; idx < (row+1)*f.RLen(); idx++ {
			f.data[idx] = blank
		}
	}

	return nil
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
	f.header.updateChanged()
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
	if fld.Type() == Numeric {
		copy(f.data[offset+fld.Len()-len(cval):], []byte(cval))
		// add spaces to the start
		for idx := offset; idx < offset+fld.Len()-len(cval); idx++ {
			f.data[idx] = blank
		}
	} else {
		copy(f.data[offset:], []byte(cval))
		// add spaces to the end
		for idx := len(cval) + offset; idx < offset+fld.Len(); idx++ {
			f.data[idx] = blank
		}
	}
	f.header.updateChanged()
	return nil
}

func (f *file) Save(w io.Writer) error {
	// write header
	buf := make([]byte, f.HLen())

	// header
	f.header.writeTo(buf)

	// fields
	for idx := range f.fields {
		f.fields[idx].descr.writeTo(buf[32+idx*32:])
	}

	// header block terminator
	buf[f.HLen()-1] = hterm

	if _, err := w.Write(buf); err != nil {
		return err
	}

	// write rows
	if _, err := w.Write(f.data); err != nil {
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
