package examples

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

// conncheck can also be used in the popular Postgres Go driver - pgx.
// The following code is commented to avoid adding a dependency to pgx itself .

//import (
//	"context"
//	"database/sql/driver"
//
//	"github.com/jackc/pgx/v5"
//	"github.com/jackc/pgx/v5/stdlib"
//	"github.com/murfffi/conncheck"
//)
//
//func pgxExample() driver.Connector {
//	return stdlib.GetConnector(pgx.ConnConfig{}, stdlib.OptionResetSession(func(_ context.Context, conn *pgx.Conn) error {
//		if conncheck.Do(conn.PgConn().Conn()) == conncheck.StatusNotOpen {
//			return driver.ErrBadConn
//		}
//		return nil
//	}))
//}
