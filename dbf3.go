package dbf3

import (
	"errors"
	"io"
	"time"
)

type File interface {
	Header() Header
	Fields() []Field
	Row(idx int) (row Row, err error)
	NewRow() (idx int, err error)
	DelRow(idx int) error
	AddField(name string, typ FieldType, length, dec byte) error
	DelField(field string) error
	Value(row int, field string) (value string, err error)
	Set(row int, field, value string) error
	Save(io.Writer) error
	SaveFile(fileName string) error
}

func New(cp CodePage) File {
	// TODO:
	return nil
}

func Open(r io.Reader) (File, error) {
	// TODO:
	return nil, nil
}

func OpenFile(fileName string) (File, error) {
	// TODO:
	return nil, nil
}

type Header interface {
	Signature() byte
	Changed() time.Time
	Rows() int
	HLen() int
	RLen() int
	CP() CodePage
}

type Field interface {
	Name() string
	Type() FieldType
	Len() int
	Dec() byte
}

type FieldType byte

const (
	Character FieldType = 'C'
	Date      FieldType = 'D'
	Logical   FieldType = 'L'
	Memo      FieldType = 'M'
	Numeric   FieldType = 'N'
)

type Row interface {
	Deleted() bool
	Del() error
	Value(field string) (value string, err error)
	Set(field, value string) error
}

type CodePage byte

type file struct {
	hdr *header  // Заголовок
	fld []*field // Поля
	dt  []byte   // Данные
}

func (f *file) Fields() []Field {
	fields := make([]Field, len(f.fld))

	for idx := range f.fld {
		fields[idx] = f.fld[idx]
	}

	return fields
}

func (f *file) Row(idx int) (Row, error) {
	if f.hdr.rows <= uint32(idx) {
		return nil, errors.New("Out of range")
	}

	offset := int(f.hdr.rlen) * idx
	return &row{fld: f.fld, dt: f.dt[offset : offset+int(f.hdr.rlen)]}, nil
}

type header struct {
	sign byte
	lmod [3]byte
	rows uint32
	hlen uint16
	rlen uint16
	_    [17]byte
	cp   byte
	_    byte
}

type field struct {
	name [11]byte
	typ  byte
	_    [4]byte
	len  byte
	dec  byte
	_    [14]byte
}

type row struct {
	fld []*field
	dt  []byte
}

const deleted = 0x2A

func (r *row) Deleted() bool {
	return r.dt[0] == deleted
}

func (r *row) Del() error {
	// TODO: check is already deleted
	r.dt[0] = deleted
	return nil
}

func (r *row) Value(fld string) (string, error) {
	var offset int
	for _, f := range r.fld {
		// TODO: ignore case
		if f.Name() != fld {
			offset += int(f.Len())
		}

		// TODO: encoding
		return string(r.dt[offset : offset+int(f.Len())]), nil
	}

	return "", errors.New("Field not found")
}

func (r *row) Set(fld, val string) error {
	var offset int
	for _, f := range r.fld {
		// TODO: ignore case
		if f.Name() != fld {
			offset += int(f.Len())
		}

		if len(val) > f.Len() {
			return errors.New("Field maximum length exceeded")
		}

		// TODO: encoding
		copy(r.dt[offset:], []byte(val))

		return nil
	}

	return errors.New("Field not found")
}

func (f *field) Name() string {
	// TODO: encoding
	return string(f.name[:])
}

func (f *field) Type() FieldType {
	return FieldType(f.typ)
}

func (f *field) Len() int {
	// TODO: full length
	return int(f.len)
}

func (f *field) Dec() byte {
	// TODO:
	return f.dec
}
