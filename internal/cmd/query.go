package cmd

import (
	"context"
	"encoding/csv"

	"gobankcli/internal/outfmt"
	"gobankcli/internal/store"
)

type QueryCmd struct {
	SQL string `arg:"" name:"sql" help:"Read-only SELECT/WITH SQL to run against the local archive."`
}

type SQLCmd QueryCmd

func (c QueryCmd) Run(ctx context.Context, app *App) error {
	return runQuery(ctx, app, c.SQL)
}

func (c SQLCmd) Run(ctx context.Context, app *App) error {
	return runQuery(ctx, app, c.SQL)
}

func runQuery(ctx context.Context, app *App, sql string) error {
	s, err := store.Open(ctx, app.Config.Paths.DB)
	if err != nil {
		return err
	}
	defer s.Close()
	result, err := s.Query(ctx, sql)
	if err != nil {
		return err
	}
	if app.OutputMode == outfmt.ModeJSON {
		return app.Out.Write(result)
	}
	return writeTSV(app.Stdout, result)
}

func writeTSV(w csvWriter, result store.QueryResult) error {
	cw := csv.NewWriter(w)
	cw.Comma = '\t'
	if err := cw.Write(result.Columns); err != nil {
		return err
	}
	for _, row := range result.Rows {
		record := make([]string, len(result.Columns))
		for i := range result.Columns {
			record[i] = row[i]
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

type csvWriter interface {
	Write([]byte) (int, error)
}
