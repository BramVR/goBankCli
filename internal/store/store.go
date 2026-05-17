package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type Store struct {
	db   *sql.DB
	path string
}

type Status struct {
	DBPath       string `json:"db_path"`
	Institutions int64  `json:"institutions"`
	Connections  int64  `json:"connections"`
	Accounts     int64  `json:"accounts"`
	Transactions int64  `json:"transactions"`
	SyncRuns     int64  `json:"sync_runs"`
}

func Open(ctx context.Context, path string) (*Store, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("db path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("mkdir db dir: %w", err)
	}
	if err := ensureDBFile(path); err != nil {
		return nil, err
	}
	dsn := sqliteDSN(path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	s := &Store{db: db, path: path}
	if err := s.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := restrictDBFiles(path); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func sqliteDSN(path string) string {
	escapedPath := (&url.URL{Path: path}).EscapedPath()
	return "file:" + escapedPath + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)"
}

func ensureDBFile(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("create sqlite file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close sqlite file: %w", err)
	}
	return restrictDBFiles(path)
}

func restrictDBFiles(path string) error {
	for _, candidate := range []string{path, path + "-shm", path + "-wal"} {
		if err := os.Chmod(candidate, 0o600); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("chmod sqlite file: %w", err)
		}
	}
	return nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

func (s *Store) migrate(ctx context.Context) error {
	var currentVersion int
	if err := s.db.QueryRowContext(ctx, "pragma user_version").Scan(&currentVersion); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if currentVersion > schemaVersion {
		return fmt.Errorf("archive schema version %d is newer than supported version %d", currentVersion, schemaVersion)
	}
	if _, err := s.db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}
	if currentVersion < schemaVersion {
		if _, err := s.db.ExecContext(ctx, fmt.Sprintf("pragma user_version = %d", schemaVersion)); err != nil {
			return fmt.Errorf("set schema version: %w", err)
		}
	}
	return nil
}

func (s *Store) Status(ctx context.Context) (Status, error) {
	status := Status{DBPath: s.path}
	counts := []struct {
		table string
		dst   *int64
	}{
		{"institutions", &status.Institutions},
		{"connections", &status.Connections},
		{"accounts", &status.Accounts},
		{"transactions", &status.Transactions},
		{"sync_runs", &status.SyncRuns},
	}
	for _, count := range counts {
		if err := s.db.QueryRowContext(ctx, "select count(*) from "+count.table).Scan(count.dst); err != nil {
			return Status{}, err
		}
	}
	return status, nil
}
