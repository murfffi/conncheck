package examples

// If you are writing a Go SQL driver that uses long-lived TCP connections to talk the database server,
// conncheck makes it easy and fast to implement the connection validation required by the
// database/sql/driver SPI.
// github.com/sclgo/impala-go is a driver that uses conncheck this way.

import (
	"context"
	"database/sql/driver"
	"net"

	"github.com/murfffi/conncheck"
)

// driverConn is an example of how to implement driver.SessionResetter in
// a database/sql driver connection using conncheck.Do.
// In production code, driverConn will also implement driver.Conn.
type driverConn struct {
	conn net.Conn
}

var _ driver.SessionResetter = (*driverConn)(nil)

func (d *driverConn) ResetSession(context.Context) error {
	if conncheck.Do(d.conn) == conncheck.StatusNotOpen {
		return driver.ErrBadConn
	}
	return nil
}
