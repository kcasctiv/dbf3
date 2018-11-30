package dbf3

import "time"

type header struct {
	sign byte
	lmod [3]byte
	rows uint32
	hlen uint16
	rlen uint16
	_    [17]byte
	cp   byte
	_    byte
}

func (h *header) Signature() byte { return h.sign }
func (h *header) Rows() int       { return int(h.rows) }
func (h *header) HLen() int       { return int(h.hlen) }
func (h *header) RLen() int       { return int(h.rlen) }
func (h *header) CP() CodePage    { return CodePage(h.cp) }

func (h *header) Changed() time.Time {
	return time.Date(
		int(h.lmod[0])+1900, time.Month(h.lmod[1]),
		int(h.lmod[2]), 0, 0, 0, 0, time.Local,
	)
}
