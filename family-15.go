package oscar

import "strconv"

func atoi(s string) int {
	i, e := strconv.Atoi(s)
	if e != nil { panic(e) }
	return i
}

type UserInfo map[string]string

func (c *Client) RequestUserInfo(username string, callback func(UserInfo)) {
	snacId := c.writer.SnacId

	c.writer.Snac(0x15, 0x02).
		Tlv(0x01).
			Len16().
				Uint32le(uint32(atoi(c.username))).Uint16le(0x07d0).Uint16le(0x0002).
					Uint16le(0x04d0).Uint32le(uint32(atoi(username))).
				LenEndUint16le().
			TlvEnd().
		SnacEnd()

	data := UserInfo(make(map[string]string))
	c.datas[snacId] = data
	c.callbacks[snacId] = callback
}

func (c *Client) UpdateUserInfo(info map[string]string, callback func(bool)) {
	snacId := c.writer.SnacId
	c.writer.SnacFlags(0x15, 0x02, 0x0001).
		Tlv(0x01).
			Len16().
				Uint32le(uint32(atoi(c.username))).Uint16le(0x07d0).Uint16le(0x0002).
					Uint16le(0x0c3a)

	for k, v := range info {
		switch k {
		case "firstname": c.writer.TlvStringZle(0x0140, v)
		case "lastname": c.writer.TlvStringZle(0x014a, v)
		case "nickname": c.writer.TlvStringZle(0x0154, v)
		case "info": c.writer.TlvStringZle(0x0258, v)
		case "auth": c.writer.TlvUint8(0x02f8, 1)
		}
	}

	c.writer.LenEndUint16le().TlvEnd().SnacEnd()

	c.callbacks[snacId] = callback
}

var subtype00c8Keys = []string{"nickname", "first name", "last name", "email", "home city", "home_state",
	"home phone", "home fax", "home address", "cell phone", "home zip code"}

func (c *Client) read1503(flags uint16, snacId uint32) {
	tlvType, _ := c.Uint16(), c.Uint16()
	if tlvType != 0x01 { panic("tlvType != 0x01") }

	_, _, dataType, _, dataSubtype := c.Uint16le(), c.Uint32le(),
		c.Uint16le(), c.Uint16le(), c.Uint16le()
	if dataType != 0x07da { panic("dataType != 0x07da") }

	flag := c.Uint8()

	switch dataSubtype {
	case 0x00c8:
		data := c.datas[snacId].(UserInfo)
		for _, key := range subtype00c8Keys {
			value := c.String(int(c.Uint16le()))
			value = windows1251ToUtf8.Conv(value)
			data[key] = value
		}

		fallthrough
	case 0x00eb, 0x00d2, 0x00f0, 0x00dc, 0x00e6, 0x00fa, 0x010e:
		if flags & 1 != 1 {
			data := c.datas[snacId].(UserInfo)
			c.datas[snacId] = data, false

			callback := c.callbacks[snacId]
			c.callbacks[snacId] = callback, false
			callback.(func(UserInfo))(data)
		}
	case 0x0c3f:
		callback := c.callbacks[snacId]
		c.callbacks[snacId] = callback, false
		callback.(func(bool))(flag == 0x0a)
	}
}

