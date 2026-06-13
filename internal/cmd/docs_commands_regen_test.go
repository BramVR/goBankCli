package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestDocsCommandsRegenIsStable(t *testing.T) {
	for _, bin := range []string{"make", "node"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("docs regen requires %q on PATH: %v", bin, err)
		}
	}

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	docsDir := filepath.Join(repoRoot, "docs", "commands")
	indexPath := filepath.Join(repoRoot, "docs", "commands.md")
	snapshot := readDocsCommandsSnapshot(t, docsDir)
	indexBefore, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read %s: %v", indexPath, err)
	}
	t.Cleanup(func() {
		restoreDocsCommandsSnapshot(t, docsDir, snapshot)
		if err := os.WriteFile(indexPath, indexBefore, 0o644); err != nil {
			t.Fatalf("restore %s: %v", indexPath, err)
		}
	})

	regen := exec.Command("make", "docs-commands")
	regen.Dir = repoRoot
	if out, err := regen.CombinedOutput(); err != nil {
		t.Fatalf("make docs-commands failed: %v\n%s", err, out)
	}

	regenerated := readDocsCommandsSnapshot(t, docsDir)
	mismatches := compareDocsCommandsSnapshots(snapshot, regenerated)
	indexAfter, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read regenerated index: %v", err)
	}
	if !bytes.Equal(indexBefore, indexAfter) {
		mismatches = append(mismatches, "docs/commands.md")
	}
	if len(mismatches) > 0 {
		sort.Strings(mismatches)
		t.Fatalf("docs regen drift; rerun make docs-commands and commit the result.\nDrifted files:\n  %s", strings.Join(mismatches, "\n  "))
	}
}

func readDocsCommandsSnapshot(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return out
	}
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		out[entry.Name()] = data
	}
	return out
}

func compareDocsCommandsSnapshots(want, got map[string][]byte) []string {
	var drift []string
	for name, wantData := range want {
		gotData, ok := got[name]
		if !ok {
			drift = append(drift, name+" (missing after regen)")
			continue
		}
		if !bytes.Equal(wantData, gotData) {
			drift = append(drift, name)
		}
	}
	for name := range got {
		if _, ok := want[name]; !ok {
			drift = append(drift, name+" (new after regen)")
		}
	}
	return drift
}

func restoreDocsCommandsSnapshot(t *testing.T, dir string, snapshot map[string][]byte) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("restore docs dir: %v", err)
	}
	for name, data := range snapshot {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatalf("restore %s: %v", name, err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s during cleanup: %v", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		if _, ok := snapshot[entry.Name()]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
			t.Fatalf("remove regenerated %s: %v", entry.Name(), err)
		}
	}
}
