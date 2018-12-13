package dbf3

type row struct {
	f   *file
	idx int
}

const (
	blank   = 0x20
	deleted = 0x2A
)

func (r *row) Deleted() bool {
	return r.f.data[r.offset()] == deleted
}

func (r *row) Del() error {
	return r.f.DelRow(r.idx)
}

func (r *row) Get(fld string) (string, error) {
	return r.f.Get(r.idx, fld)
}

func (r *row) Set(fld, val string) error {
	return r.f.Set(r.idx, fld, val)
}

func (r *row) offset() int {
	return r.f.RLen() * r.idx
}
