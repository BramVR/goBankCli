package outfmt

import (
	"bytes"
	"strings"
	"testing"
)

func TestPlainWriterRendersTopLevelSlice(t *testing.T) {
	var buf bytes.Buffer
	err := New(&buf, ModePlain).Write([]struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{{ID: "one", Name: "First"}})
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "id: one\n") || !strings.Contains(got, "name: First\n") {
		t.Fatalf("plain slice output = %q", got)
	}
	if strings.Contains(got, "{") || strings.Contains(got, "}") {
		t.Fatalf("plain slice output uses Go syntax: %q", got)
	}
}

func TestPlainWriterRendersNestedSlice(t *testing.T) {
	var buf bytes.Buffer
	report := struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
		Count int `json:"count"`
	}{
		Items: []struct {
			ID string `json:"id"`
		}{{ID: "one"}},
		Count: 1,
	}
	if err := New(&buf, ModePlain).Write(report); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "items[0].id: one\n") || !strings.Contains(got, "count: 1\n") {
		t.Fatalf("plain nested slice output = %q", got)
	}
	if strings.Contains(got, "{") || strings.Contains(got, "}") {
		t.Fatalf("plain nested slice output uses Go syntax: %q", got)
	}
}
