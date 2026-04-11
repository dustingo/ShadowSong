package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/game-ops/ai-alert-system/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const roleAuditQuery = `
SELECT role, COUNT(*) AS count
FROM users
GROUP BY role
ORDER BY role
`

type roleBucket struct {
	Role  string
	Count int64
}

type rowScanner interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (rowScanner, error)
}

type closeableQueryer interface {
	queryer
	Close() error
}

type sqlConnQueryer struct {
	db *sql.DB
	conn *sql.Conn
}

const setReadOnlySQL = "SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY"

func (q *sqlConnQueryer) QueryContext(ctx context.Context, query string, args ...any) (rowScanner, error) {
	rows, err := q.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (q *sqlConnQueryer) Close() error {
	if q.conn != nil {
		if err := q.conn.Close(); err != nil {
			_ = q.db.Close()
			return err
		}
	}
	if q.db != nil {
		return q.db.Close()
	}
	return nil
}

func main() {
	os.Exit(run(context.Background(), os.Stdout, os.Stderr))
}

func run(ctx context.Context, stdout, stderr io.Writer) int {
	dbCfg := config.LoadDatabaseConfig()
	db, err := openReadOnlyDB(ctx, dbCfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to connect to database: %v\n", err)
		return 1
	}
	defer db.Close()

	return runAudit(ctx, db, stdout, stderr)
}

type sessionExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func openReadOnlyDB(ctx context.Context, cfg config.DatabaseConfig) (closeableQueryer, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	if err := setSessionReadOnly(ctx, conn); err != nil {
		conn.Close()
		db.Close()
		return nil, err
	}

	return &sqlConnQueryer{db: db, conn: conn}, nil
}

func setSessionReadOnly(ctx context.Context, execer sessionExecer) error {
	_, err := execer.ExecContext(ctx, setReadOnlySQL)
	return err
}

func runAudit(ctx context.Context, db queryer, stdout, stderr io.Writer) int {
	buckets, err := auditRoleCounts(ctx, db)
	if err != nil {
		fmt.Fprintf(stderr, "role audit failed: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, "Persisted role counts:")
	for _, bucket := range buckets {
		fmt.Fprintf(stdout, "- %s: %d\n", bucket.Role, bucket.Count)
	}

	unsupported := unsupportedBuckets(buckets)
	if len(unsupported) == 0 {
		fmt.Fprintf(stdout, "Audit passed: only supported roles found (%s)\n", strings.Join(authz.SupportedRoles(), ", "))
		return 0
	}

	fmt.Fprintln(stderr, "Unsupported persisted roles detected:")
	for _, bucket := range unsupported {
		fmt.Fprintf(stderr, "- %s: %d\n", bucket.Role, bucket.Count)
	}
	return 1
}

func auditRoleCounts(ctx context.Context, db queryer) ([]roleBucket, error) {
	rows, err := db.QueryContext(ctx, roleAuditQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buckets []roleBucket
	for rows.Next() {
		var bucket roleBucket
		if err := rows.Scan(&bucket.Role, &bucket.Count); err != nil {
			return nil, err
		}
		buckets = append(buckets, bucket)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return buckets, nil
}

func unsupportedBuckets(buckets []roleBucket) []roleBucket {
	var unsupported []roleBucket
	for _, bucket := range buckets {
		if !authz.IsSupportedRole(bucket.Role) {
			unsupported = append(unsupported, bucket)
		}
	}

	sort.Slice(unsupported, func(i, j int) bool {
		return unsupported[i].Role < unsupported[j].Role
	})

	return unsupported
}
