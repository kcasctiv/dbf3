package dbf3

import (
	"encoding/binary"
	"time"
)

type header struct {
	SG byte     // signature
	LM [3]byte  // last modification date
	RW uint32   // rows count
	HL uint16   // header length
	RL uint16   // row length
	_  [17]byte // reserved
	LD byte     // language driver ID
	_  [2]byte  // reserved
}

func readHeader(buf []byte) header {
	var h header
	h.SG = buf[0]
	copy(h.LM[:], buf[1:4])
	h.RW = binary.LittleEndian.Uint32(buf[4:8])
	h.HL = binary.LittleEndian.Uint16(buf[8:10])
	h.RL = binary.LittleEndian.Uint16(buf[10:12])
	h.LD = buf[29]
	return h
}

func (h *header) changed() time.Time {
	return time.Date(
		int(h.LM[0])+1900, time.Month(h.LM[1]),
		int(h.LM[2]), 0, 0, 0, 0, time.Local,
	)
}

func (h *header) updateChanged() {
	now := time.Now()
	h.LM[0] = byte(now.Year() - 1900)
	h.LM[1] = byte(now.Month())
	h.LM[2] = byte(now.Day())
}
