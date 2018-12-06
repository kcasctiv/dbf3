package dbf3

import (
	"bytes"
	"encoding/binary"
	"strings"
)

type field struct {
	dt     fieldData
	name   string // field name
	idx    int    // field index
	offset int    // field offset inside row
	len    int    // full length
}

type fieldData struct {
	NM [11]byte // name
	TP byte     // type
	_  [4]byte  // reserved
	LN byte     // length
	DC byte     // decimal count
	_  [14]byte // reserved
}

func newField(dt fieldData, idx int, offset int) *field {
	var length uint16
	switch FieldType(dt.TP) {
	case Character:
		// up to 64kb
		buf := bytes.NewBuffer([]byte{dt.LN, dt.DC})
		binary.Read(buf, binary.LittleEndian, &length)
	default:
		length = uint16(dt.LN)
	}
	return &field{
		dt:     dt,
		name:   strings.Trim(string(dt.NM[:]), string([]byte{0})),
		idx:    idx,
		offset: offset,
		len:    int(length),
	}
}

func (f *field) Name() string    { return f.name }
func (f *field) Type() FieldType { return FieldType(f.dt.TP) }
func (f *field) Len() int        { return f.len }
func (f *field) Dec() byte       { return f.dt.DC }
