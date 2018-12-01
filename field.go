package dbf3

import (
	"bytes"
	"encoding/binary"
	"strings"
)

type field struct {
	name [11]byte
	typ  byte
	_    [4]byte
	len  byte
	dec  byte
	_    [14]byte
}

func (f *field) Name() string {
	return strings.Trim(string(f.name[:]), string([]byte{0}))
}

func (f *field) Type() FieldType {
	return FieldType(f.typ)
}

func (f *field) Len() int {
	switch FieldType(f.typ) {
	case Character:
		// up to 64kb
		buf := bytes.NewBuffer([]byte{f.len, f.dec})
		var length uint16
		binary.Read(buf, binary.BigEndian, &length)
		return int(length)
	default:
		return int(f.len)
	}
}

func (f *field) Dec() byte {
	return f.dec
}
