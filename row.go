package dbf3

type row struct {
	f   *file
	idx int
}

const valid = 0x20
const deleted = 0x2A

func (r *row) Deleted() bool {
	return r.f.dt[r.offset()] == deleted
}

func (r *row) Del() error {
	return r.f.DelRow(r.idx)
}

func (r *row) Value(fld string) (string, error) {
	return r.f.Value(r.idx, fld)
}

func (r *row) Set(fld, val string) error {
	return r.f.Set(r.idx, fld, val)
}

func (r *row) offset() int {
	return int(r.f.hdr.rlen) * r.idx
}
