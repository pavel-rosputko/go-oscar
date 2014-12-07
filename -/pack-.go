package oscar

import (
	"bytes"
	"reflect"
)

type Packer struct {
	bytes.Buffer
}

func (p *Packer) Uint8(v uint8) {
	p.WriteByte(v)
}

func (p *Packer) Uint16(v uint16) {
	p.WriteByte(byte(v >> 8))
	p.WriteByte(byte(v))
}

func (p *Packer) Uint32(v uint32) {
	p.WriteByte(byte(v >> 24))
	p.WriteByte(byte(v >> 16))
	p.WriteByte(byte(v >> 8))
	p.WriteByte(byte(v))
}

func (p *Packer) Values(values ...interface{}) {
	for _, v := range values {
		switch t := reflect.Typeof(v).(type) {
		case *reflect.UintType:
			switch t.Size() {
			case 4:
				i := v.(uint32)
				p.WriteByte(byte(i >> 24))
				p.WriteByte(byte(i >> 16))
				p.WriteByte(byte(i >> 8))
				p.WriteByte(byte(i))
			}
		}
	}
}

func Pack(values ...interface{}) []byte {
	p := new(Packer)
	p.Values(values...)
	return p.Bytes()
}
