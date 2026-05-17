package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type QueryResult struct {
	Columns []string   `json:"columns"`
	Rows    [][]string `json:"rows"`
}

func (s *Store) Query(ctx context.Context, query string) (QueryResult, error) {
	query, err := readOnlySQL(query)
	if err != nil {
		return QueryResult{}, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return QueryResult{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "pragma query_only = on"); err != nil {
		return QueryResult{}, err
	}
	defer tx.ExecContext(context.Background(), "pragma query_only = off")
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return QueryResult{}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return QueryResult{}, err
	}
	result := QueryResult{Columns: columns, Rows: [][]string{}}
	values := make([]any, len(columns))
	scan := make([]any, len(columns))
	for i := range values {
		scan[i] = &values[i]
	}
	for rows.Next() {
		if err := rows.Scan(scan...); err != nil {
			return QueryResult{}, err
		}
		row := make([]string, len(columns))
		for i := range columns {
			row[i] = queryValueString(values[i])
		}
		result.Rows = append(result.Rows, row)
	}
	return result, rows.Err()
}

func readOnlySQL(query string) (string, error) {
	query = strings.TrimSpace(query)
	var err error
	query, err = trimSQLTerminator(query)
	if err != nil {
		return "", err
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return "", errors.New("query is required")
	}
	token := firstSQLToken(query)
	switch token {
	case "select", "with":
		return query, nil
	default:
		return "", fmt.Errorf("only read-only select queries are allowed, got %q", token)
	}
}

func trimSQLTerminator(query string) (string, error) {
	inSingle := false
	inDouble := false
	inLineComment := false
	inBlockComment := false
	for i := 0; i < len(query); i++ {
		ch := query[i]
		next := byte(0)
		if i+1 < len(query) {
			next = query[i+1]
		}
		switch {
		case inLineComment:
			if ch == '\n' {
				inLineComment = false
			}
		case inBlockComment:
			if ch == '*' && next == '/' {
				inBlockComment = false
				i++
			}
		case inSingle:
			if ch == '\'' && next == '\'' {
				i++
			} else if ch == '\'' {
				inSingle = false
			}
		case inDouble:
			if ch == '"' && next == '"' {
				i++
			} else if ch == '"' {
				inDouble = false
			}
		case ch == '-' && next == '-':
			inLineComment = true
			i++
		case ch == '/' && next == '*':
			inBlockComment = true
			i++
		case ch == '\'':
			inSingle = true
		case ch == '"':
			inDouble = true
		case ch == ';':
			if firstSQLToken(query[i+1:]) != "" {
				return "", errors.New("query must contain one statement")
			}
			return strings.TrimSpace(query[:i]), nil
		}
	}
	return query, nil
}

func firstSQLToken(query string) string {
	query = strings.TrimSpace(query)
	for {
		switch {
		case strings.HasPrefix(query, "--"):
			if idx := strings.IndexByte(query, '\n'); idx >= 0 {
				query = strings.TrimSpace(query[idx+1:])
				continue
			}
			return ""
		case strings.HasPrefix(query, "/*"):
			if idx := strings.Index(query, "*/"); idx >= 0 {
				query = strings.TrimSpace(query[idx+2:])
				continue
			}
			return ""
		}
		break
	}
	fields := strings.Fields(query)
	if len(fields) == 0 {
		return ""
	}
	return strings.ToLower(fields[0])
}

func queryValueString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case []byte:
		return string(v)
	case time.Time:
		return v.UTC().Format(time.RFC3339Nano)
	default:
		return fmt.Sprint(v)
	}
}
