package dbf3

import "errors"

type row struct {
	fld []*field
	dt  []byte
}

const valid = 0x20
const deleted = 0x2A

func (r *row) Deleted() bool {
	return r.dt[0] == deleted
}

func (r *row) Del() error {
	if r.dt[0] == deleted {
		return errors.New("Already deleted")
	}
	r.dt[0] = deleted
	return nil
}

func (r *row) Value(fld string) (string, error) {
	var offset int
	for _, f := range r.fld {
		// TODO: ignore case
		if f.Name() != fld {
			offset += int(f.Len())
		}

		// TODO: encoding
		return string(r.dt[offset : offset+int(f.Len())]), nil
	}

	return "", errors.New("Field not found")
}

func (r *row) Set(fld, val string) error {
	var offset int
	for _, f := range r.fld {
		// TODO: ignore case
		if f.Name() != fld {
			offset += int(f.Len())
		}

		if len(val) > f.Len() {
			return errors.New("Field maximum length exceeded")
		}

		// TODO: encoding
		copy(r.dt[offset:], []byte(val))

		return nil
	}

	return errors.New("Field not found")
}
