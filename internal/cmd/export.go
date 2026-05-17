package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gobankcli/internal/config"
	"gobankcli/internal/csvexport"
	"gobankcli/internal/outfmt"
	"gobankcli/internal/store"
)

type ExportCmd struct {
	From      string `help:"Start booking date, inclusive, as YYYY-MM-DD."`
	To        string `help:"End booking date, inclusive, as YYYY-MM-DD."`
	AccountID string `name:"account" help:"Restrict export to one local account ID."`
	Out       string `help:"CSV output path. Use - for stdout." type:"path"`
}

type exportReport struct {
	OutputPath string `json:"output_path"`
	Rows       int    `json:"rows"`
}

func (c ExportCmd) Run(ctx context.Context, app *App) error {
	from, err := parseOptionalDate(c.From, "from")
	if err != nil {
		return err
	}
	to, err := parseOptionalDate(c.To, "to")
	if err != nil {
		return err
	}
	if from != nil && to != nil && from.After(*to) {
		return errors.New("from date must be on or before to date")
	}

	s, err := store.Open(ctx, app.Config.Paths.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	rows, err := s.ListTransactions(ctx, store.TransactionFilter{
		From:      from,
		To:        to,
		AccountID: c.AccountID,
	})
	if err != nil {
		return err
	}

	outPath := c.Out
	if outPath == "" {
		outPath = filepath.Join(app.Config.Paths.Exports, "normalized.csv")
	}
	outPath = config.ExpandPath(outPath)
	if outPath == "-" {
		if app.OutputMode != outfmt.ModeHuman {
			return errors.New("--json/--plain cannot be used when exporting CSV to stdout")
		}
		return csvexport.Write(app.Stdout, rows)
	}

	if err := writeCSVFile(outPath, rows); err != nil {
		return err
	}
	return app.Out.Write(exportReport{OutputPath: outPath, Rows: len(rows)})
}

func parseOptionalDate(value, name string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, fmt.Errorf("%s must use YYYY-MM-DD: %w", name, err)
	}
	return &t, nil
}

func writeCSVFile(path string, rows []csvexport.Row) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	return csvexport.Write(f, rows)
}
