package dbf3

import "time"

type header struct {
	SG byte     // signature
	LM [3]byte  // last modification date
	RW uint32   // rows count
	HL uint16   // header length
	RL uint16   // row length
	_  [17]byte // reserved
	CP byte     // code page
	_  [2]byte  // reserved
}

func (h *header) Signature() byte    { return h.SG }
func (h *header) Rows() int          { return int(h.RW) }
func (h *header) HLen() int          { return int(h.HL) }
func (h *header) RLen() int          { return int(h.RL) }
func (h *header) CodePage() CodePage { return CodePage(h.CP) }

func (h *header) Changed() time.Time {
	return time.Date(
		int(h.LM[0])+1900, time.Month(h.LM[1]),
		int(h.LM[2]), 0, 0, 0, 0, time.Local,
	)
}
