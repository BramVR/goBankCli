package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobankcli/internal/store"
)

func TestExportRejectsArchiveDBOutputPath(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "archive.db")
	s, err := store.Open(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err = Run(ctx, []string{"--db", dbPath, "export", "--out", dbPath}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "archive database") {
		t.Fatalf("Run error = %v, want archive database output rejection", err)
	}
	after, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, before) {
		t.Fatal("archive database changed after rejected export")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestExportRejectsArchiveDBAliasOutputPath(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	realDBPath := filepath.Join(dir, "archive.db")
	dbLinkPath := filepath.Join(dir, "configured.db")
	s, err := store.Open(ctx, realDBPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realDBPath, dbLinkPath); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(realDBPath)
	if err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err = Run(ctx, []string{"--db", dbLinkPath, "export", "--out", realDBPath}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "archive database") {
		t.Fatalf("Run error = %v, want archive database alias output rejection", err)
	}
	after, err := os.ReadFile(realDBPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, before) {
		t.Fatal("archive database changed after rejected alias export")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestExportRejectsMissingArchiveDBAliasOutputPath(t *testing.T) {
	dir := t.TempDir()
	realDBPath := filepath.Join(dir, "archive.db")
	dbLinkPath := filepath.Join(dir, "configured.db")
	if err := os.Symlink(realDBPath, dbLinkPath); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--db", dbLinkPath, "export", "--out", realDBPath}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "archive database") {
		t.Fatalf("Run error = %v, want archive database alias output rejection", err)
	}
	if _, err := os.Stat(realDBPath); !os.IsNotExist(err) {
		t.Fatalf("archive database should not be created before rejected export: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestExportRejectsArchiveSidecarOutputPath(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "archive.db")
	outPath := dbPath + "-wal"

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--db", dbPath, "export", "--out", outPath}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "archive database") {
		t.Fatalf("Run error = %v, want archive database sidecar output rejection", err)
	}
	if _, err := os.Stat(outPath); !os.IsNotExist(err) {
		t.Fatalf("archive sidecar should not be created before rejected export: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestExportRejectsMissingArchiveSidecarThroughSymlinkedParent(t *testing.T) {
	realDir := t.TempDir()
	linkRoot := t.TempDir()
	aliasDir := filepath.Join(linkRoot, "alias")
	if err := os.Symlink(realDir, aliasDir); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(realDir, "archive.db")
	outPath := filepath.Join(aliasDir, "archive.db-wal")

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--db", dbPath, "export", "--out", outPath}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "archive database") {
		t.Fatalf("Run error = %v, want archive database sidecar output rejection", err)
	}
	if _, err := os.Stat(filepath.Join(realDir, "archive.db-wal")); !os.IsNotExist(err) {
		t.Fatalf("archive sidecar should not be created before rejected export: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestExportRejectsExistingSymlinkOutputPath(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "archive.db")
	targetPath := filepath.Join(dir, "target.csv")
	linkPath := filepath.Join(dir, "export.csv")
	before := []byte("keep me\n")
	if err := os.WriteFile(targetPath, before, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--db", dbPath, "export", "--out", linkPath}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("Run error = %v, want symlink output rejection", err)
	}
	after, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, before) {
		t.Fatal("symlink target changed after rejected export")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestExportWritesPrivateCSVFile(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "archive.db")
	outPath := filepath.Join(dir, "exports", "normalized.csv")
	if err := os.MkdirAll(filepath.Dir(outPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outPath, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--db", dbPath, "export", "--out", outPath}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(b), "date,value_date,account_id,iban,institution,counterparty_name,counterparty_account,description,amount,currency,transaction_id,provider,category\n") {
		t.Fatalf("export CSV header mismatch:\n%s", b)
	}
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("export permissions = %o, want 600", got)
	}
	if !strings.Contains(stdout.String(), "rows: 0") {
		t.Fatalf("export report missing row count:\n%s", stdout.String())
	}
}

func TestExportStdoutCSVUnchanged(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "archive.db")

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--db", dbPath, "export", "--out", "-"}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stdout.String(), "date,value_date,account_id,iban,institution,counterparty_name,counterparty_account,description,amount,currency,transaction_id,provider,category\n") {
		t.Fatalf("stdout CSV header mismatch:\n%s", stdout.String())
	}
	if strings.Contains(stdout.String(), "rows:") {
		t.Fatalf("stdout export should not include report:\n%s", stdout.String())
	}
}
