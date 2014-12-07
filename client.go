package oscar

import (
	"net"
	"os"

	"g/log"
)

var Logger log.Logger = new(log.NullLogger)

func debug(v ...interface{}) {
	Logger.Debug(v...)
}

func debugf(f string, v ...interface{}) {
	Logger.Debugf(f, v...)
}

type Listener interface {
	Ready()
	Message(username, text string)
	Subscription(username string)
}

type Client struct {
	Reader
	writer	Writer
	Listener

	username	string
	password	string

	callbacks	map[uint32]interface{}
	datas		map[uint32]interface{}
}

func New(username, password string, listener Listener) *Client {
	return &Client{
		username: username,
		password: password,
		Listener: listener,
		callbacks: make(map[uint32]interface{}),
		datas: make(map[uint32]interface{})}
}

func (c *Client) readFlap() {
	c.Read(6)

	c.Skip(1)
	c.Uint8()
	c.Skip(2)
	flapLen := c.Uint16()

	c.Read(int(flapLen))
}

func (c *Client) readSnac() {
	c.readFlap()
	c.Uint16()
	c.Uint16()
	flags := c.Uint16()
	c.Uint32()

	if flags == 1 << 15 { c.Skip(int(c.Uint16())) }
}

func (c *Client) readSnacType() (snacType uint32) {
	c.readFlap()
	snacType = c.Uint32()
	flags := c.Uint16()
	c.Uint32()

	if flags == 1 << 15 { c.Skip(int(c.Uint16())) }
	return
}

var roastData = []byte{0xf3, 0x26, 0x81, 0xc4, 0x39, 0x86, 0xdb, 0x92,
	0x71, 0xa3, 0xb9, 0xe6, 0x53, 0x7a, 0x95, 0x7c}

func roast(s string) []byte {
	b := []byte(s)
	for i := 0; i < len(s); i++ {
		b[i] ^= roastData[i % len(roastData)]
	}

	return b
}

func (c *Client) writeIdent() {
	c.writer.Flap(0x01).
		Uint32(1).
		TlvString(0x01, c.username).
		TlvBytes(0x02, roast(c.password)).
		TlvString(0x03, "Go OSCAR").
		TlvUint16(0x16, 0xbeef).
		TlvUint16(0x17, 0x1).
		TlvUint16(0x18, 0x0).
		TlvUint16(0x19, 0x0).
		TlvUint16(0x1a, 0x0).
		TlvUint32(0x14, 0x1).
		TlvString(0x0f, "en").
		TlvString(0x0e, "us").
		FlapEnd()
}

type AuthError uint16

func (e AuthError) String() string {
	return "auth"
}

func (c *Client) readAuthResp() (string, string) {
	c.readFlap()

	var addr, cookie string
	var error uint16
	for c.HasRest() {
		tlvType, tlvLen := c.Uint16(), c.Uint16()

		switch tlvType {
		case 0x05: addr = c.String(int(tlvLen))
		case 0x06: cookie = c.String(int(tlvLen))
		case 0x08: error = c.Uint16()
		default: c.Skip(int(tlvLen))
		}
	}

	// TODO what if addr == ""
	if error != 0 { panic(AuthError(error)) }

	return addr, cookie
}

func (c *Client) dial(addr string) {
	debug("dial: addr =", addr)
	conn, err := net.Dial("tcp", "", addr)
	if err != nil { panic(err) }
	c.Reader = MakeReader(conn)
	c.writer = MakeWriter(conn)
}

func (c *Client) read0103() {
	if c.readSnacType() != 0x00010003 { panic("unexpected snac") }
}

func (c *Client) read0115() {
	if c.readSnacType() != 0x00010015 { panic("unexpected snac") }
}

var FamilyVersions = map[uint16]uint16{0x01: 3, 0x02: 1, 0x03: 1, 0x04: 1, 0x06: 1, 0x09: 1,
	0x0a: 1, 0x0b: 1, 0x13: 2, 0x15: 1}
// if 0x13 is set to 1 then snac: 0x13, 0x1c is not sent !

func (c *Client) write0117() {
	c.writer.Snac(0x01, 0x17)
	for f, v := range FamilyVersions {
		c.writer.Uint16(f).Uint16(v)
	}
	c.writer.SnacEnd()
}

func (c *Client) read0118() {
	if c.readSnacType() != 0x00010018 { panic("unexpected snac") }
}

func (c *Client) read0113() {
	if c.readSnacType() != 0x00010013 { panic("unexpected snac") }
}

func (c *Client) write0102() {
	c.writer.Snac(0x01, 0x02)
	for f, v := range FamilyVersions {
		c.writer.Uint16(f).Uint16(v).Uint16(0x0110).Uint16(0x047b)
	}
	c.writer.SnacEnd()
}

func (c *Client) read0b02() {
	if c.readSnacType() != 0x000b0002 { panic("unexpected snac") }
}

type DisconnectError uint16

func (e DisconnectError) String() string {
	return "disconnect"
}

func (c *Client) handleFlap04() {
	var code uint16
	for c.HasRest() {
		tlvType, tlvLen := c.Uint16(), c.Uint16()

		switch tlvType {
		case 0x09:
			code = c.Uint16()
		default:
			c.Skip(int(tlvLen))
		}
	}

	panic(DisconnectError(code))
}

const AuthAddr = "login.icq.com:5190"

func (c *Client) Run() (e os.Error) {
	defer func() {
		e = recover().(os.Error)
	}()

	c.dial(AuthAddr)

	c.readFlap()
	c.writeIdent()

	addr, cookie := c.readAuthResp()

	c.dial(addr)

	c.readFlap()
	c.writer.Flap(0x01).
		Uint32(1).
		TlvString(0x06, cookie).
		FlapEnd()

	c.read0103()

	c.read0115()

	c.write0117()
	c.read0118()

	c.read0113()

	c.write0102()

	c.read0b02()

	c.Listener.Ready()

	for {
		c.Read(6)

		c.Skip(1)
		ch := c.Uint8()
		c.Skip(2)
		flapLen := c.Uint16()

		c.Read(int(flapLen))

		switch ch {
		case 0x02:
			snacType, flags, snacId := c.Uint32(), c.Uint16(), c.Uint32()
			if flags == 1 << 15 { c.Skip(int(c.Uint16())) }

			switch snacType {
			case 0x00040007:
				c.read0407()
			case 0x0013001c:
				c.read131c()
			case 0x00150003:
				c.read1503(flags, snacId)
			}
		case 0x04:
			c.handleFlap04()
		}
	}

	return
}
