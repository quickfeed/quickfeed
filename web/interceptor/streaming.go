package interceptor

import (
	"connectrpc.com/connect"
)

// checkedConn wraps a streaming handler connection and a check function.
type checkedConn struct {
	connect.StreamingHandlerConn
	check   func(any) error
	checked bool
}

// Receive receives a message from the connection, and runs the check function
// on the first message received. After the first message, Receive behaves
// like the underlying connection's Receive method. Receive is called from the
// handler's goroutine, so checked needs no synchronization.
func (c *checkedConn) Receive(msg any) error {
	if err := c.StreamingHandlerConn.Receive(msg); err != nil {
		return err
	}
	if c.checked {
		return nil
	}
	c.checked = true
	return c.check(msg)
}
