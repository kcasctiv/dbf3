package dbf3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"time"
)

// File presents DBF file interface
type File interface {
	Header() Header
	Fields() []Field
	Row(idx int) (row Row, err error)
	NewRow() (idx int, err error)
	DelRow(idx int) error
	Deleted(idx int) (bool, error)
	AddField(name string, typ FieldType, length, dec byte) error
	DelField(field string) error
	Value(row int, field string) (value string, err error)
	Set(row int, field, value string) error
	Save(io.Writer) error
	SaveFile(fileName string) error
}

// New creates new empty DBF file
func New(cp CodePage) File {
	now := time.Now()
	return &file{
		hdr: &header{
			SG: 0x03, // dbase 3 without DBT
			HL: 33,   // header + terminator
			RL: 1,    // no fields + deletion flag
			CP: byte(cp),
			LM: [3]byte{
				byte(now.Year() - 1900),
				byte(now.Month()),
				byte(now.Day()),
			},
		},
		dt: []byte{0x1a}, // EOF only
	}
}

// Open opens DBF from reader
func Open(r io.Reader) (File, error) {
	var hdr header
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}

	flds := make([]*field, (hdr.HLen()-33)/32)
	for idx := range flds {
		var fld field
		err := binary.Read(r, binary.LittleEndian, &fld)
		if err != nil {
			return nil, err
		}
		flds[idx] = &fld
	}

	var term byte
	err := binary.Read(r, binary.LittleEndian, &term)
	if err != nil {
		return nil, err
	}
	if term != 0x0D {
		return nil, errors.New("not expected header terminator")
	}

	var buf [512]byte
	data := bytes.NewBuffer(nil)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if _, err := data.Write(buf[:n]); err != nil {
			return nil, err
		}
	}

	explen := hdr.RLen()*hdr.Rows() + 1
	if data.Len() != explen {
		return nil, errors.New("not expected length of records block")
	}

	return &file{
		hdr: &hdr,
		fld: flds,
		dt:  data.Bytes(),
	}, nil
}

// OpenFile opens DBF from file
func OpenFile(fileName string) (File, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Open(file)
}

// Header presents DBF header interface
type Header interface {
	Signature() byte
	Changed() time.Time
	Rows() int
	HLen() int
	RLen() int
	CodePage() CodePage
}

// Field presents DBF field descriptor
type Field interface {
	Name() string
	Type() FieldType
	Len() int
	Dec() byte
}

// FieldType presents type of DBF field
type FieldType byte

// Supported field types
const (
	Character FieldType = 'C'
	Date      FieldType = 'D'
	Logical   FieldType = 'L'
	Numeric   FieldType = 'N'
)

// Row presents DBF row interface
type Row interface {
	Deleted() bool
	Del() error
	Value(field string) (value string, err error)
	Set(field, value string) error
}

// CodePage presents DBF code page
type CodePage byte

// Supported code pages
const (
	CP866 CodePage = 0x65
)
