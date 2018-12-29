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
	Changed() time.Time
	Rows() int
	HLen() int
	RLen() int
	Lang() LangID
	SetLang(lang LangID)
	Fields() []Field
	HasField(field string) bool
	Row(idx int) (row Row, err error)
	NewRow() (idx int, err error)
	DelRow(idx int) error
	Deleted(idx int) (bool, error)
	AddField(name string, typ FieldType, length, dec byte) error
	DelField(field string) error
	Get(row int, field string) (value string, err error)
	Set(row int, field, value string) error
	Save(io.Writer) error
	SaveFile(fileName string) error
}

// New creates new empty DBF file
func New(lang LangID) File {
	now := time.Now()
	return &file{
		header: header{
			signature: 0x03, // dbase 3 without DBT
			hlen:      33,   // header + terminator
			rlen:      1,    // no fields + deletion flag
			lang:      byte(lang),
			changed: [3]byte{
				byte(now.Year() - 1900),
				byte(now.Month()),
				byte(now.Day()),
			},
		},
		data:      []byte{eof}, // EOF only
		fieldsIdx: make(map[string]int),
		converter: newCharmapsTextConverter(lang),
	}
}

// Open opens DBF from reader
func Open(rd io.Reader) (File, error) {
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

	return &file{
		header:    hdr,
		fields:    fields,
		data:      buf,
		fieldsIdx: fieldsIdx,
		converter: newCharmapsTextConverter(LangID(hdr.lang)),
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
	Get(field string) (value string, err error)
	Set(field, value string) error
}

// LangID presents DBF language driver ID
type LangID byte

// CodePage returns code page number for language driver
func (l LangID) CodePage() string {
	return codepages[l]
}

// Supported language driver ids
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
	Lang19      LangID = 0x13 // Japanese Shift-JIS
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
	Lang77      LangID = 0x4D // Chinese GBK (PRC)
	Lang78      LangID = 0x4E // Korean (ANSI/OEM)
	Lang79      LangID = 0x4F // Chinese Big5 (Taiwan)
	Lang80      LangID = 0x50 // Thai (ANSI/OEM)
	Lang87      LangID = 0x57 // ANSI
	Lang88      LangID = 0x58 // Western European ANSI
	Lang89      LangID = 0x59 // Spanish ANSI
	Lang100     LangID = 0x64 // Eastern European MS-DOS
	Lang101     LangID = 0x65 // Russian MS-DOS
	Lang102     LangID = 0x66 // Nordic MS-DOS
	Lang103     LangID = 0x67 // Icelandic MS-DOS
	Lang104     LangID = 0x68 // Kamenicky (Czech) MS-DOS
	Lang105     LangID = 0x69 // Mazovia (Polish) MS-DOS
	Lang106     LangID = 0x6A // Greek MS-DOS (437G)
	Lang107     LangID = 0x6B // Turkish MS-DOS
	Lang108     LangID = 0x6C // French-Canadian MS-DOS
	Lang120     LangID = 0x78 // Taiwan Big 5
	Lang121     LangID = 0x79 // Hangul (Wansung)
	Lang122     LangID = 0x7A // PRC GBK
	Lang123     LangID = 0x7B // Japanese Shift-JIS
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
