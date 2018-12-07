package dbf3

import (
	"encoding/binary"
	"time"
)

type header struct {
	signature byte     // signature
	changed   [3]byte  // last modification date
	rows      uint32   // rows count
	hlen      uint16   // header length
	rlen      uint16   // row length
	_         [17]byte // reserved
	lang      byte     // language driver ID
	_         [2]byte  // reserved
}

func readHeader(buf []byte) header {
	var h header
	h.signature = buf[0]
	copy(h.changed[:], buf[1:])
	h.rows = binary.LittleEndian.Uint32(buf[4:])
	h.hlen = binary.LittleEndian.Uint16(buf[8:])
	h.rlen = binary.LittleEndian.Uint16(buf[10:])
	h.lang = buf[29]
	return h
}

func (h *header) writeTo(buf []byte) {
	buf[0] = h.signature
	copy(buf[1:], h.changed[:])
	binary.LittleEndian.PutUint32(buf[4:], h.rows)
	binary.LittleEndian.PutUint16(buf[8:], h.hlen)
	binary.LittleEndian.PutUint16(buf[10:], h.rlen)
	buf[29] = h.lang
}

func (h *header) changedTime() time.Time {
	return time.Date(
		int(h.changed[0])+1900, time.Month(h.changed[1]),
		int(h.changed[2]), 0, 0, 0, 0, time.Local,
	)
}

func (h *header) updateChanged() {
	now := time.Now()
	h.changed[0] = byte(now.Year() - 1900)
	h.changed[1] = byte(now.Month())
	h.changed[2] = byte(now.Day())
}
