// Package dbf3 intended for reading and writing xbase/dbase III files
package dbf3

import (
	"bufio"
	"errors"
	"io"
	"os"
	"time"
)

// File presents DBF file interface
type File interface {
	// Changed returns date of last file change
	Changed() time.Time
	// Rows returns rows count
	Rows() int
	// HLen returns length of file header
	HLen() int
	// RLen returns length of file row
	RLen() int
	// Lang returns language driver identifier
	Lang() LangID
	// SetLang sets language driver of file
	SetLang(lang LangID)
	// Fields returns fields list
	Fields() []Field
	// HasField checks if file contains field with specified name
	HasField(field string) bool
	// Row returns row with specified index
	Row(idx int) (row Row, err error)
	// NewRow creates new row and returns its index
	NewRow() (idx int, err error)
	// DelRow marks row with specified index as deleted
	DelRow(idx int) error
	// Deleted checks if row with specified index marked as deleted
	Deleted(idx int) (deleted bool, err error)
	// Pack removes rows marked as deleted
	Pack() error
	// AddField adds new field in file (in the end of row)
	AddField(name string, typ FieldType, length, dec byte) error
	// DelField deletes field from file (with all values of that field in all rows)
	DelField(field string) error
	// Get returns field value from row with specified index
	Get(row int, field string) (value string, err error)
	// Set sets field value in row with specified index
	Set(row int, field, value string) error
	// Save writes dbf into specified io.Writer
	Save(w io.Writer) error
	// SaveFile saves dbf into file with specified name
	SaveFile(fileName string) error
}

type options struct {
	lang     LangID
	convCtor TextConverterCtor
}

func newDefaultOptions() *options {
	return &options{
		lang:     LangDefault,
		convCtor: CharmapsTextConverter,
	}
}

// Option presents DBF option
type Option func(*options)

// WithLang presents language option
func WithLang(lang LangID) func(*options) {
	return func(o *options) {
		o.lang = lang
	}
}

// WithTextConverter presents text converter option
func WithTextConverter(ctor TextConverterCtor) func(*options) {
	return func(o *options) {
		if ctor != nil {
			o.convCtor = ctor
		}
	}
}

// New creates new empty DBF file
func New(opts ...Option) File {
	now := time.Now()

	o := newDefaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &file{
		header: header{
			signature: 0x03, // dbase 3 without DBT
			hlen:      33,   // header + terminator
			rlen:      1,    // no fields + deletion flag
			lang:      byte(o.lang),
			changed: [3]byte{
				byte(now.Year() - 1900),
				byte(now.Month()),
				byte(now.Day()),
			},
		},
		data:          []byte{eof}, // EOF only
		fieldsIdx:     make(map[string]int),
		converterCtor: o.convCtor,
		converter:     o.convCtor(o.lang),
	}
}

// Open opens DBF from reader
func Open(rd io.Reader, opts ...Option) (File, error) {
	r := bufio.NewReader(rd)

	buf := make([]byte, 32)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	hdr := readHeader(buf)

	fields := make([]*field, (int(hdr.hlen)-33)/32)
	fieldsIdx := make(map[string]int)
	offset := 1 // fields starts after deletion flag
	var fd fieldDescr

	buf = make([]byte, int(hdr.hlen)-32)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	for idx := range fields {
		fd.readFrom(buf[idx*32:])

		fields[idx] = newField(fd, idx, offset)
		fieldsIdx[fields[idx].Name()] = idx
		offset += fields[idx].Len()
	}

	term := buf[len(fields)*32]
	if term != hterm {
		return nil, errors.New("not expected header terminator")
	}

	buf = make([]byte, int(hdr.rlen)*int(hdr.rows)+1)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	o := newDefaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	if o.lang != LangDefault {
		hdr.lang = byte(o.lang)
	}

	return &file{
		header:        hdr,
		fields:        fields,
		data:          buf,
		fieldsIdx:     fieldsIdx,
		converterCtor: o.convCtor,
		converter:     o.convCtor(LangID(hdr.lang)),
	}, nil
}

// OpenFile opens DBF from file
func OpenFile(fileName string, opts ...Option) (File, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Open(file, opts...)
}

// Field presents DBF field descriptor
type Field interface {
	// Name returns name of the field
	Name() string
	// Type returns type of the field
	Type() FieldType
	// Len returns length of the field
	Len() int
	// Dec returns decimals count of the field
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
	// Deleted checks if row marked as deleted
	Deleted() bool
	// Del marks row as deleted
	Del() error
	// Get returns field value
	Get(field string) (value string, err error)
	// Set sets field value
	Set(field, value string) error
}

// LangID presents DBF language driver ID
type LangID byte

// CodePage returns code page number for language driver
func (l LangID) CodePage() string {
	return codepages[l]
}

// Supported language driver ids
//
// List of drivers and mapped code pages taken
// from http://www.autopark.ru/ASBProgrammerGuide/DBFSTRUC.HT
// and maybe not full and/or not correct
const (
	LangDefault LangID = 0x00 // Use default driver
	Lang1       LangID = 0x01 // US MS-DOS
	Lang2       LangID = 0x02 // International MS-DOS
	Lang3       LangID = 0x03 // Windows ANSI Latin I
	Lang4       LangID = 0x04 // Standard Macintosh
	Lang8       LangID = 0x08 // Danish OEM
	Lang9       LangID = 0x09 // Dutch OEM
	Lang10      LangID = 0x0A // Dutch OEM*
	Lang11      LangID = 0x0B // Finnish OEM
	Lang13      LangID = 0x0D // French OEM
	Lang14      LangID = 0x0E // French OEM*
	Lang15      LangID = 0x0F // German OEM
	Lang16      LangID = 0x10 // German OEM*
	Lang17      LangID = 0x11 // Italian OEM
	Lang18      LangID = 0x12 // Italian OEM*
	Lang19      LangID = 0x13 // Japanese Shift-JIS 				(not implemented)
	Lang20      LangID = 0x14 // Spanish OEM*
	Lang21      LangID = 0x15 // Swedish OEM
	Lang22      LangID = 0x16 // Swedish OEM*
	Lang23      LangID = 0x17 // Norwegian OEM
	Lang24      LangID = 0x18 // Spanish OEM
	Lang25      LangID = 0x19 // English OEM (Great Britain)
	Lang26      LangID = 0x1A // English OEM (Great Britain)*
	Lang27      LangID = 0x1B // English OEM (US)
	Lang28      LangID = 0x1C // French OEM (Canada)
	Lang29      LangID = 0x1D // French OEM*
	Lang31      LangID = 0x1F // Czech OEM
	Lang34      LangID = 0x22 // Hungarian OEM
	Lang35      LangID = 0x23 // Polish OEM
	Lang36      LangID = 0x24 // Portuguese OEM
	Lang37      LangID = 0x25 // Portuguese OEM*
	Lang38      LangID = 0x26 // Russian OEM
	Lang55      LangID = 0x37 // English OEM (US)*
	Lang64      LangID = 0x40 // Romanian OEM
	Lang77      LangID = 0x4D // Chinese GBK (PRC) 					(not implemented)
	Lang78      LangID = 0x4E // Korean (ANSI/OEM) 					(not implemented)
	Lang79      LangID = 0x4F // Chinese Big5 (Taiwan) 				(not implemented)
	Lang80      LangID = 0x50 // Thai (ANSI/OEM)
	Lang87      LangID = 0x57 // ANSI
	Lang88      LangID = 0x58 // Western European ANSI
	Lang89      LangID = 0x59 // Spanish ANSI
	Lang100     LangID = 0x64 // Eastern European MS-DOS
	Lang101     LangID = 0x65 // Russian MS-DOS
	Lang102     LangID = 0x66 // Nordic MS-DOS
	Lang103     LangID = 0x67 // Icelandic MS-DOS
	Lang104     LangID = 0x68 // Kamenicky (Czech) MS-DOS 			(not implemented)
	Lang105     LangID = 0x69 // Mazovia (Polish) MS-DOS 			(not implemented)
	Lang106     LangID = 0x6A // Greek MS-DOS (437G)
	Lang107     LangID = 0x6B // Turkish MS-DOS
	Lang108     LangID = 0x6C // French-Canadian MS-DOS
	Lang120     LangID = 0x78 // Taiwan Big 5 						(not implemented)
	Lang121     LangID = 0x79 // Hangul (Wansung)					(not implemented)
	Lang122     LangID = 0x7A // PRC GBK							(not implemented)
	Lang123     LangID = 0x7B // Japanese Shift-JIS					(not implemented)
	Lang124     LangID = 0x7C // Thai Windows/MS–DOS
	Lang134     LangID = 0x86 // Greek OEM
	Lang135     LangID = 0x87 // Slovenian OEM
	Lang136     LangID = 0x88 // Turkish OEM
	Lang150     LangID = 0x96 // Russian Macintosh
	Lang151     LangID = 0x97 // Eastern European Macintosh
	Lang152     LangID = 0x98 // Greek Macintosh
	Lang200     LangID = 0xC8 // Eastern European Windows
	Lang201     LangID = 0xC9 // Russian Windows
	Lang202     LangID = 0xCA // Turkish Windows
	Lang203     LangID = 0xCB // Greek Windows
	Lang204     LangID = 0xCC // Baltic Windows
)
