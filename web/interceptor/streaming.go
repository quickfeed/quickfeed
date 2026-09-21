package interceptor

import (
	"connectrpc.com/connect"
)

// checkedConn applies check to the first message a streaming handler receives.
// For a server stream that message is the request the handler is about to act
// on, which is otherwise out of reach: connect hands a streaming interceptor
// the connection, not the request, so access control and validation can only
// run at the point the handler reads it.
//
// Receive is called from the handler's goroutine, so checked needs no
// synchronization.
type checkedConn struct {
	connect.StreamingHandlerConn
	check   func(any) error
	checked bool
}

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
