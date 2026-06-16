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

	outPath := c.Out
	if outPath == "" {
		outPath = filepath.Join(app.Config.Paths.Exports, "normalized.csv")
	}
	outPath = config.ExpandPath(outPath)
	if outPath == "-" {
		if app.OutputMode != outfmt.ModeHuman {
			return errors.New("--json/--plain cannot be used when exporting CSV to stdout")
		}
	} else if err := validateCSVOutputPath(outPath, app.Config.Paths.DB); err != nil {
		return err
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

	if outPath == "-" {
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

func samePath(a, b string) bool {
	absA, err := filepath.Abs(a)
	if err != nil {
		absA = a
	}
	absB, err := filepath.Abs(b)
	if err != nil {
		absB = b
	}
	return filepath.Clean(absA) == filepath.Clean(absB)
}

func validateCSVOutputPath(outPath, dbPath string) error {
	if samePath(outPath, dbPath) {
		return fmt.Errorf("CSV output path must not be the archive database: %s", outPath)
	}
	info, err := os.Lstat(outPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("CSV output path must not be an existing symlink: %s", outPath)
	}
	return nil
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
	if err := f.Chmod(0o600); err != nil {
		return err
	}
	return csvexport.Write(f, rows)
}
