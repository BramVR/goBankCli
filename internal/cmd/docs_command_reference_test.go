package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestDocsCommandReferenceEmitsCLIMetadataWithoutSecrets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GOBANKCLI_GOCARDLESS_SECRET_ID", "fixture-docs-id")
	t.Setenv("GOBANKCLI_GOCARDLESS_SECRET_KEY", "fixture-docs-key")

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"docs-command-reference"}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	out := stdout.String()
	for _, secret := range []string{"fixture-docs-id", "fixture-docs-key"} {
		if strings.Contains(out, secret) {
			t.Fatalf("metadata leaked secret %q:\n%s", secret, out)
		}
	}

	var doc commandReferenceDocument
	if err := json.Unmarshal(stdout.Bytes(), &doc); err != nil {
		t.Fatalf("metadata is not JSON: %v\n%s", err, out)
	}
	if doc.Version != 1 || doc.Binary != "gobankcli" {
		t.Fatalf("unexpected metadata header: %+v", doc)
	}
	query := findReferenceCommand(doc.Commands, "query")
	if query == nil {
		t.Fatalf("query command missing from metadata: %+v", doc.Commands)
	}
	if query.Usage != "gobankcli query <sql> [flags]" || query.PositionalArgs != "<sql>" {
		t.Fatalf("query usage/args mismatch: %+v", query)
	}
	if !hasReferenceFlag(query.Flags, "json", "bool", "false") {
		t.Fatalf("query flags missing global --json with false default: %+v", query.Flags)
	}

	connect := findReferenceCommand(doc.Commands, "connect")
	if connect == nil {
		t.Fatalf("connect command missing from metadata: %+v", doc.Commands)
	}
	if !hasReferenceFlag(connect.Flags, "callback-timeout", "duration", "5m") {
		t.Fatalf("connect flags missing --callback-timeout default: %+v", connect.Flags)
	}
	if findReferenceCommand(doc.Commands, "docs-command-reference") != nil {
		t.Fatalf("hidden docs command should not be in generated metadata: %+v", doc.Commands)
	}
}

func findReferenceCommand(commands []commandReferenceCommand, name string) *commandReferenceCommand {
	for i := range commands {
		if commands[i].Name == name {
			return &commands[i]
		}
	}
	return nil
}

func hasReferenceFlag(flags []commandReferenceFlag, name, typ, def string) bool {
	for _, flag := range flags {
		if flag.Name == name && flag.Type == typ && flag.Default == def {
			return true
		}
	}
	return false
}
