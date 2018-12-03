package dbf3

import (
	"bytes"
	"encoding/binary"
	"strings"
)

type field struct {
	NM [11]byte // name
	TP byte     // type
	_  [4]byte  // reserved
	LN byte     // length
	DC byte     // decimal count
	_  [14]byte // reserved
}

func (f *field) Name() string {
	return strings.Trim(string(f.NM[:]), string([]byte{0}))
}

func (f *field) Type() FieldType {
	return FieldType(f.TP)
}

func (f *field) Len() int {
	switch FieldType(f.TP) {
	case Character:
		// up to 64kb
		buf := bytes.NewBuffer([]byte{f.LN, f.DC})
		var length uint16
		binary.Read(buf, binary.BigEndian, &length)
		return int(length)
	default:
		return int(f.LN)
	}
}

func (f *field) Dec() byte {
	return f.DC
}
