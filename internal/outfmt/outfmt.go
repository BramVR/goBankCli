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
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		return writePlainSlice(w, "", rv)
	}
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
		valueField := reflect.Indirect(rv.Field(i))
		if valueField.Kind() == reflect.Slice || valueField.Kind() == reflect.Array {
			if err := writePlainSlice(w, name, valueField); err != nil {
				return err
			}
			continue
		}
		value := rv.Field(i).Interface()
		if _, err := fmt.Fprintf(w, "%s: %v\n", name, value); err != nil {
			return err
		}
	}
	return nil
}

func writePlainSlice(w io.Writer, prefix string, rv reflect.Value) error {
	for i := 0; i < rv.Len(); i++ {
		item := reflect.Indirect(rv.Index(i))
		if item.Kind() != reflect.Struct {
			name := prefix
			if name == "" {
				name = "item"
			}
			if _, err := fmt.Fprintf(w, "%s[%d]: %v\n", name, i, rv.Index(i).Interface()); err != nil {
				return err
			}
			continue
		}
		rt := item.Type()
		for j := 0; j < item.NumField(); j++ {
			field := rt.Field(j)
			if field.PkgPath != "" {
				continue
			}
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name == "" || name == "-" {
				name = strings.ToLower(field.Name)
			}
			if prefix != "" {
				name = fmt.Sprintf("%s[%d].%s", prefix, i, name)
			}
			if _, err := fmt.Fprintf(w, "%s: %v\n", name, item.Field(j).Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}
