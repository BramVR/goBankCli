package outfmt

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type Mode string

const (
	ModeHuman Mode = "human"
	ModeJSON  Mode = "json"
	ModePlain Mode = "plain"
)

type Writer struct {
	w    io.Writer
	mode Mode
}

func New(w io.Writer, mode Mode) *Writer {
	return &Writer{w: w, mode: mode}
}

func (w *Writer) Write(v any) error {
	switch w.mode {
	case ModeJSON:
		enc := json.NewEncoder(w.w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	case ModePlain:
		return writePlain(w.w, v)
	default:
		return writePlain(w.w, v)
	}
}

func writePlain(w io.Writer, v any) error {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if rv.Kind() != reflect.Struct {
		_, err := fmt.Fprintln(w, v)
		return err
	}
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			name = strings.ToLower(field.Name)
		}
		value := rv.Field(i).Interface()
		if _, err := fmt.Fprintf(w, "%s: %v\n", name, value); err != nil {
			return err
		}
	}
	return nil
}
