package examples

// conncheck fits great with any TCP connection pool.
// This example configures yudhasubki/netpool with conncheck.

// The following code is commented to avoid keeping a dependency to pgx itself .

/*
import (
	"context"
	"errors"
	"net"

	"github.com/murfffi/conncheck"
	"github.com/yudhasubki/netpool"
)

func netpoolExample() (*netpool.Netpool, error) {
	brokenErr := errors.New("broken connection")
	pool, err := netpool.New(func(ctx context.Context) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, "tcp", "127.0.0.1:8080")
	}, netpool.Config{
		MaxPool: 100,
		MinPool: 10,
		HealthCheck: func(conn net.Conn) error {
			if conncheck.Do(conn) == conncheck.StatusNotOpen {
				return brokenErr
			}
			return nil
		},
	})
	return pool, err
}
*/
