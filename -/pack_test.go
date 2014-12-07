package oscar

import (
	"bytes"
	"testing"
	"encoding/binary"
)

func Test(t *testing.T) {
}

func BenchmarkEncBin(bm *testing.B) {
	b := bytes.NewBuffer(nil)
	for i := 0; i < bm.N; i++ {
		binary.Write(b, binary.LittleEndian, int32(i))
	}
	println(len(b.Bytes()))
}

func BenchmarkWrite(bm *testing.B) {
	b := bytes.NewBuffer(nil)
	for i := 0; i < bm.N; i++ {
		b.WriteByte(byte(i >> 24))
		b.WriteByte(byte(i >> 16))
		b.WriteByte(byte(i >> 8))
		b.WriteByte(byte(i))
	}

	println(len(b.Bytes()))
}

func BenchmarkPack(bm *testing.B) {
	var b Packer
	for i := 0; i < bm.N; i++ {
		b.Uint32(uint32(i))
	}
	println(len(b.Bytes()))
}

func BenchmarkPackValue(bm *testing.B) {
	var p Packer
	for i := 0; i < bm.N; i++ {
		p.Values(uint32(i))
	}
	println(len(p.Bytes()))
}

func (p *Packer) Pack1(value interface{}) {
	p.Uint32(value.(uint32))
}

func BenchmarkPackTypeAssert(bm *testing.B) {
	var p Packer
	for i := 0; i < bm.N; i++ {
		p.Pack1(uint32(i))
	}
	println(len(p.Bytes()))
}

func BenchmarkPackUint32(bm *testing.B) {
	var p Packer
	for i := 0; i < bm.N; i++ {
		v := uint32(i)
		p.Uint32(v)
		p.Uint32(v)
		p.Uint32(v)
		p.Uint32(v)
	}
	println(len(p.Bytes()))
}

func BenchmarkPackValues(bm *testing.B) {
	var p Packer
	for i := 0; i < bm.N; i++ {
		v := uint32(i)
		p.Values(v, v, v, v)
	}
	println(len(p.Bytes()))
}

func Benchmark(bm *testing.B) {
	var p Packer
	for i := 0; i < bm.N; i++ {
		v := uint32(i)
		p.Write(Pack(v, v, v, v))
	}
	println(len(p.Bytes()))
}
