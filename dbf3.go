package dbf3

import (
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
			sign: 0x03, // dbase 3 without DBT
			hlen: 33,   // header + terminator
			rlen: 1,    // no fields + deletion flag
			cp:   byte(cp),
			lmod: [3]byte{
				byte(now.Year() - 1900),
				byte(now.Month()),
				byte(now.Day()),
			},
		},
	}
}

// Open opens DBF from reader
func Open(r io.Reader) (File, error) {
	// TODO:
	return nil, nil
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
	CP() CodePage
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
	Memo      FieldType = 'M'
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
