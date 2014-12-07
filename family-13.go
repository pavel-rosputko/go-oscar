package oscar

func (c *Client) read131c() {
	username := c.String(int(c.Uint8()))
	c.Listener.Subscription(username)
}
