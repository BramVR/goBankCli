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
	return canonicalPath(a) == canonicalPath(b)
}

func canonicalPath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		if abs, err := filepath.Abs(resolved); err == nil {
			return filepath.Clean(abs)
		}
		return filepath.Clean(resolved)
	}
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	if resolvedDir, err := filepath.EvalSymlinks(dir); err == nil {
		path = filepath.Join(resolvedDir, base)
	}
	if abs, err := filepath.Abs(path); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}

func validateCSVOutputPath(outPath, dbPath string) error {
	if isArchiveOutputPath(outPath, dbPath) {
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

func isArchiveOutputPath(outPath, dbPath string) bool {
	for _, protectedPath := range archiveOutputPaths(dbPath) {
		if samePath(outPath, protectedPath) || sameExistingFile(outPath, protectedPath) {
			return true
		}
	}
	return false
}

func archiveOutputPaths(dbPath string) []string {
	bases := []string{dbPath}
	if resolvedDir, err := filepath.EvalSymlinks(filepath.Dir(dbPath)); err == nil {
		bases = appendUniquePath(bases, filepath.Join(resolvedDir, filepath.Base(dbPath)))
	}
	if resolvedDBPath, ok := resolveSymlinkTarget(dbPath); ok {
		bases = appendUniquePath(bases, resolvedDBPath)
	}

	var paths []string
	for _, base := range bases {
		paths = appendUniquePath(paths, base)
		paths = appendUniquePath(paths, base+"-wal")
		paths = appendUniquePath(paths, base+"-shm")
	}
	return paths
}

func resolveSymlinkTarget(path string) (string, bool) {
	current := path
	for range 16 {
		target, err := os.Readlink(current)
		if err != nil {
			return current, current != path
		}
		if filepath.IsAbs(target) {
			current = target
		} else {
			current = filepath.Join(filepath.Dir(current), target)
		}
	}
	return current, current != path
}

func appendUniquePath(paths []string, path string) []string {
	for _, existing := range paths {
		if samePath(existing, path) {
			return paths
		}
	}
	return append(paths, path)
}

func sameExistingFile(a, b string) bool {
	infoA, err := os.Stat(a)
	if err != nil {
		return false
	}
	infoB, err := os.Stat(b)
	if err != nil {
		return false
	}
	return os.SameFile(infoA, infoB)
}

func writeCSVFile(path string, rows []csvexport.Row) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := f.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()
	if err := f.Chmod(0o600); err != nil {
		_ = f.Close()
		return err
	}
	if err := csvexport.Write(f, rows); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}
