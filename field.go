package dbf3

import (
	"encoding/binary"
	"strings"
)

type field struct {
	descr  fieldDescr
	name   string // field name
	idx    int    // field index
	offset int    // field offset inside row
	len    int    // full length
}

func (f *field) Name() string    { return f.name }
func (f *field) Type() FieldType { return FieldType(f.descr.typ) }
func (f *field) Len() int        { return f.len }
func (f *field) Dec() byte       { return f.descr.dec }

type fieldDescr struct {
	name [11]byte // name
	typ  byte     // type
	_    [4]byte  // reserved
	len  byte     // length
	dec  byte     // decimal count
	_    [14]byte // reserved
}

func (fd *fieldDescr) readFrom(buf []byte) {
	copy(fd.name[:], buf[:11])
	fd.typ = buf[11]
	fd.len = buf[16]
	fd.dec = buf[17]
}

func (fd fieldDescr) writeTo(buf []byte) {
	copy(buf, fd.name[:])
	buf[11] = fd.typ
	buf[16] = fd.len
	buf[17] = fd.dec
}

func newField(dt fieldDescr, idx int, offset int) *field {
	var length uint16
	switch FieldType(dt.typ) {
	case Character:
		// up to 64kb
		// (theoretically. in fact up to 32kb,
		// because max length of row is 64kb
		// and we need at least 1 byte for deletion flag)
		length = binary.LittleEndian.Uint16([]byte{dt.len, dt.dec})
	default:
		length = uint16(dt.len)
	}
	return &field{
		descr:  dt,
		name:   strings.Trim(string(dt.name[:]), string([]byte{0})),
		idx:    idx,
		offset: offset,
		len:    int(length),
	}
}
