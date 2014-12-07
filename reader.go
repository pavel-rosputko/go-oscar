package oscar

import (
	"io"
)

type Reader struct {
	io.Reader

	b	[]byte
	i	int
}

func MakeReader(ioReader io.Reader) Reader {
	return Reader{Reader: ioReader}
}

func (r *Reader) Read(l int) {
	debug("Reader#Read: l =", l)

	r.b = make([]byte, l)
	_, e := io.ReadFull(r.Reader, r.b)
	if e != nil { panic(e) }

	debug("Reader#Read: r.b =", r.b)

	r.i = 0
}

func (r *Reader) Uint8() (v uint8) {
	v = uint8(r.b[r.i])
	r.i += 1
	return
}

func (r *Reader) Uint16() (v uint16) {
	v = uint16(r.b[r.i]) << 8 | uint16(r.b[r.i + 1])
	r.i += 2
	return
}

func (r *Reader) Uint32() (v uint32) {
	v = uint32(r.b[r.i]) << 24 | uint32(r.b[r.i + 1]) << 16 |
		uint32(r.b[r.i + 2]) << 8 | uint32(r.b[r.i + 3])
	r.i += 4
	return
}

func (r *Reader) Uint16le() (v uint16) {
	v = uint16(r.b[r.i + 1]) << 8 | uint16(r.b[r.i])
	r.i += 2
	return
}

func (r *Reader) Uint32le() (v uint32) {
	v = uint32(r.b[r.i + 3]) << 24 | uint32(r.b[r.i + 2]) << 16 |
		uint32(r.b[r.i + 1]) << 8 | uint32(r.b[r.i])
	r.i += 4
	return
}

func (r *Reader) Skip(l int) {
	r.i += l
}

func (r *Reader) RestLen() int {
	return len(r.b) - r.i
}

func (r *Reader) HasRest() bool {
	return r.i < len(r.b)
}

func (r *Reader) Bytes(l int) (v []byte) {
	v = r.b[r.i : r.i + l]
	r.i += l
	return
}

func (r *Reader) String(l int) (v string) {
	v = string(r.b[r.i : r.i + l])
	r.i += l
	return
}
