package examples

// conncheck can be used in the popular Postgres Go driver - pgx.
// By default, pgx does not check connection health before picking it up from the pool.
// Instead, it provides hooks for such checks.
// The same approach would work for any Go SQL driver or DB pool that uses long-lived TCP connections to
// talk to the database and exposes hooks for validating connections.
//
// The following code is commented to avoid keeping a dependency to pgx itself .

/*
import (
	"context"
	"database/sql"
	"database/sql/driver"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/murfffi/conncheck"
)

// pgxStdLibExample shows how to add connection check to the pgx stdlib API
func pgxStdLibExample() {
	connector := stdlib.GetConnector(pgx.ConnConfig{}, stdlib.OptionResetSession(func(_ context.Context, conn *pgx.Conn) error {
		if conncheck.Do(conn.PgConn().Conn()) == conncheck.StatusNotOpen {
			return driver.ErrBadConn
		}
		return nil
	}))
	var db *sql.DB = sql.OpenDB(connector)

	// work with the database as needed
	_ = db.Ping()
}

// pgxPoolExample shows how to add connection check to the pgxpool API
func pgxPoolExample() error {
	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	config.PrepareConn = func(ctx context.Context, conn *pgx.Conn) (bool, error) {
		return conncheck.Do(conn.PgConn().Conn()) != conncheck.StatusNotOpen, nil
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return err
	}

	// work with the database as needed
	err = pool.Ping(context.Background())

	return err
}
*/
