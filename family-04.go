package oscar

import (
	"bytes"
	"rand"

	"g/iconv"
)

const (
	utf8 = "{0946134E-4C7F-11D1-8222-444553540000}"
	rtf = "{97B12751-243C-4334-AD22-D6ABF73F1492}"
)

var icqRelay = []byte("\x09\x46\x13\x49\x4c\x7f\x11\xd1\x82\x22\x44\x45\x53\x54\x00\x00")
var zeroPlugin = make([]byte, 16)

var windows1251ToUtf8 = iconv.Open("utf-8", "windows-1251")
var utf16beToUtf8 = iconv.Open("utf-8", "utf-16be")
var utf8ToUtf16be = iconv.Open("utf-16be", "utf-8")

func (c *Client) read0407() {
	c.Skip(8)
	ch := c.Uint16()
	username := c.String(int(c.Uint8()))
	c.Skip(2)
	count := int(c.Uint16())
	for i := 0; i < count; i++ {
		c.Skip(2)
		tlvLen := c.Uint16()
		c.Skip(int(tlvLen))
	}

	var text string
	switch ch {
	case 0x01:
		tlvType, tlvLen := c.Uint16(), c.Uint16()
		if tlvType != 0x02 { panic("tlvType != 0x02") }

		var encoding uint16
		for c.HasRest() {
			tlvType, tlvLen = c.Uint16(), c.Uint16()
			switch tlvType {
			case 0x0101:
				encoding = c.Uint16()
				c.Skip(2)
				text = c.String(int(tlvLen) - 4)
			default:
				c.Skip(int(tlvLen))
			}
		}

		if encoding == 2 {
			text = utf16beToUtf8.Conv(text)
		} else {
			text = windows1251ToUtf8.Conv(text)
		}
	case 0x02:
		tlvType, _ := c.Uint16(), c.Uint16()
		if tlvType != 0x05 { panic("tlvType != 0x05") }

		c.Skip(2)
		c.Skip(8)

		capability := c.Bytes(16)

		var messageType uint8
		var uuid string

		if bytes.Equal(capability, icqRelay) {
			tlvType, tlvLen := c.Uint16(), c.Uint16()
			for tlvType != 0x2711 {
				c.Skip(int(tlvLen))
				tlvType, tlvLen = c.Uint16(), c.Uint16()
			}

			c.Skip(4)
			plugin := c.Bytes(16)
			c.Skip(25)

			if bytes.Equal(plugin, zeroPlugin) {
				messageType = c.Uint8()
				c.Skip(5)
				text = c.String(int(c.Uint16le()))
				// TODO unpack Z*

				if messageType == 0x01 {
					c.Skip(8)
					uuidLen := int(c.Uint16le())
					if c.RestLen() >= uuidLen {
						uuid = c.String(uuidLen)
					}
				}
			}
		}

		if text != "" && messageType == 0x01 {
			switch uuid {
			case utf8:
			case rtf:
				text = ""
			default:
				text = windows1251ToUtf8.Conv(text)
			}
		}
	}

	if text != "" {
		c.Listener.Message(username, text)
	}
}

func (c *Client) SendMessage(username, text string) {
	text = utf8ToUtf16be.Conv(text)

	c.writer.Snac(0x04, 0x06).
		Uint32(0).Uint32(uint32(rand.Int31n(10000000))).
			Uint16(1).Uint8(uint8(len(username))).String(username).
		Tlv(0x02).
			TlvUint16(0x0501, 0x0001).
			Tlv(0x0101).Uint16(0x0002).Uint16(0x0000).String(text).TlvEnd().
			TlvEnd().
		SnacEnd()
}
