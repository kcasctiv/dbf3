package dbf3

type field struct {
	name [11]byte
	typ  byte
	_    [4]byte
	len  byte
	dec  byte
	_    [14]byte
}

func (f *field) Name() string {
	// TODO: encoding
	return string(f.name[:])
}

func (f *field) Type() FieldType {
	return FieldType(f.typ)
}

func (f *field) Len() int {
	// TODO: full length
	return int(f.len)
}

func (f *field) Dec() byte {
	// TODO:
	return f.dec
}
