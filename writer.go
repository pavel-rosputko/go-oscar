package oscar

import (
	"bytes"
	"io"
	"rand"
)

type intStack []int

func (s *intStack) Push(v int) {
	*s = append(*s, v)
}

func (s *intStack) Pop() (v int) {
	v, *s = (*s)[len(*s) - 1], (*s)[:len(*s) - 1]
	return
}

type Writer struct {
	*bytes.Buffer

	writer io.Writer
	FlapId	uint16
	SnacId	uint32
	stack	intStack
}

func MakeWriter(ioWriter io.Writer) Writer {
	return Writer{
		Buffer: bytes.NewBuffer(nil),
		writer: ioWriter,
		FlapId: uint16(rand.Int()),
		SnacId: uint32(rand.Int())}
}

func (w *Writer) Uint8(v uint8) *Writer {
	w.WriteByte(v)
	return w
}

func (w *Writer) Uint16(v uint16) *Writer {
	w.WriteByte(byte(v >> 8))
	w.WriteByte(byte(v))
	return w
}

func (w *Writer) Uint32(v uint32) *Writer {
	w.WriteByte(byte(v >> 24))
	w.WriteByte(byte(v >> 16))
	w.WriteByte(byte(v >> 8))
	w.WriteByte(byte(v))
	return w
}

func (w *Writer) Uint16le(v uint16) *Writer {
	w.WriteByte(byte(v))
	w.WriteByte(byte(v >> 8))
	return w
}

func (w *Writer) Uint32le(v uint32) *Writer {
	w.WriteByte(byte(v))
	w.WriteByte(byte(v >> 8))
	w.WriteByte(byte(v >> 16))
	w.WriteByte(byte(v >> 24))
	return w
}

func (w *Writer) Bytes(bytes []byte) *Writer {
	w.Write(bytes)
	return w
}

func (w *Writer) String(s string) *Writer {
	w.WriteString(s)
	return w
}

func (w *Writer) NextFlapId() (id uint16) {
	id = w.FlapId
	w.FlapId++
	return
}

func (w *Writer) NextSnacId() (id uint32) {
	id = w.SnacId
	w.SnacId++
	return
}

func (w *Writer) Len16() *Writer {
	w.Uint16(0)
	w.stack.Push(len(w.Buffer.Bytes()))
	return w
}

func (w *Writer) LenEndUint16() *Writer {
	i := w.stack.Pop()
	b := w.Buffer.Bytes()[i - 2:]
	l := len(w.Buffer.Bytes()) - i
	b[0], b[1] = byte(l >> 8), byte(l)
	return w
}

func (w *Writer) LenEndUint16le() *Writer {
	i := w.stack.Pop()
	b := w.Buffer.Bytes()[i - 2:]
	l := len(w.Buffer.Bytes()) - i
	b[0], b[1] = byte(l), byte(l >> 8)
	return w
}

func (w *Writer) Flap(flapType uint8) *Writer {
	return w.Uint8('*').Uint8(flapType).Uint16(w.NextFlapId()).Len16()
}

func (w *Writer) FlapEnd() {
	w.LenEndUint16()

	debug("Writer#FlapEnd: bytes =", w.Buffer.Bytes())

	w.WriteTo(w.writer)
}

func (w *Writer) Snac(family, subtype uint16) *Writer {
	return w.Flap(0x02).
		Uint16(family).Uint16(subtype).Uint16(0).Uint32(w.NextSnacId())
}

func (w *Writer) SnacFlags(family, subtype, flags uint16) *Writer {
	return w.Flap(0x02).
		Uint16(family).Uint16(subtype).Uint16(flags).Uint32(w.NextSnacId())
}

func (w *Writer) SnacEnd() {
	w.FlapEnd()
}

func (w *Writer) Tlv(tlvType uint16) *Writer {
	return w.Uint16(tlvType).Len16()
}

func (w *Writer) TlvEnd() *Writer {
	return w.LenEndUint16()
}

func (w *Writer) TlvString(t uint16, v string) *Writer {
	return w.Uint16(t).Uint16(uint16(len(v))).String(v)
}

func (w *Writer) TlvBytes(t uint16, v []byte) *Writer {
	return w.Uint16(t).Uint16(uint16(len(v))).Bytes(v)
}

func (w *Writer) TlvUint8(t uint16, v uint8) *Writer {
	return w.Uint16(t).Uint16(1).Uint8(v)
}

func (w *Writer) TlvUint16(t, v uint16) *Writer {
	return w.Uint16(t).Uint16(2).Uint16(v)
}

func (w *Writer) TlvUint32(t uint16, v uint32) *Writer {
	return w.Uint16(t).Uint16(4).Uint32(v)
}

func (w *Writer) TlvStringZle(t uint16, v string) *Writer {
	return w.Uint16le(t).Len16().
		Uint16le(uint16(len(v))).String(v).Uint8(0).
		LenEndUint16le()
}
