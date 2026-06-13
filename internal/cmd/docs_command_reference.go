package cmd

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/alecthomas/kong"
)

const commandReferenceVersion = 1

type DocsCommandReferenceCmd struct{}

type commandReferenceDocument struct {
	Version  int                       `json:"version"`
	Binary   string                    `json:"binary"`
	Commands []commandReferenceCommand `json:"commands"`
}

type commandReferenceCommand struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Usage          string                 `json:"usage"`
	PositionalArgs string                 `json:"positional_args,omitempty"`
	Flags          []commandReferenceFlag `json:"flags"`
}

type commandReferenceFlag struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     string `json:"default"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

func emitCommandReference(parser *kong.Kong, stdout io.Writer) error {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(commandReferenceFromModel(parser.Model))
}

func commandReferenceFromModel(model *kong.Application) commandReferenceDocument {
	commands := make([]commandReferenceCommand, 0)
	for _, leaf := range model.Leaves(true) {
		if leaf.Type != kong.CommandNode || leaf.Hidden || leaf.Name == "docs-command-reference" {
			continue
		}
		commands = append(commands, commandReferenceFromNode(model.Name, leaf))
	}
	return commandReferenceDocument{
		Version:  commandReferenceVersion,
		Binary:   model.Name,
		Commands: commands,
	}
}

func commandReferenceFromNode(binary string, node *kong.Node) commandReferenceCommand {
	flags := make([]commandReferenceFlag, 0)
	for _, group := range node.AllFlags(true) {
		for _, flag := range group {
			if flag.Name == "help" {
				continue
			}
			flags = append(flags, commandReferenceFlag{
				Name:        flag.Name,
				Type:        commandReferenceFlagType(flag.Value),
				Default:     commandReferenceDefault(flag.Value),
				Description: trimSentence(flag.Help),
				Required:    flag.Required,
			})
		}
	}
	positionals := make([]string, 0, len(node.Positional))
	for _, arg := range node.Positional {
		positionals = append(positionals, arg.Summary())
	}
	return commandReferenceCommand{
		Name:           node.Name,
		Description:    trimSentence(firstString(node.Detail, node.Help)),
		Usage:          strings.TrimSpace(binary + " " + node.Summary()),
		PositionalArgs: strings.Join(positionals, " "),
		Flags:          flags,
	}
}

func commandReferenceFlagType(value *kong.Value) string {
	if value == nil || !value.Target.IsValid() {
		return "string"
	}
	if value.Target.Type() == reflect.TypeOf(time.Duration(0)) {
		return "duration"
	}
	switch value.Target.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "uint"
	case reflect.Float32, reflect.Float64:
		return "float"
	default:
		return "string"
	}
}

func commandReferenceDefault(value *kong.Value) string {
	if value == nil {
		return ""
	}
	if value.HasDefault {
		return value.Default
	}
	if value.Target.IsValid() && value.Target.Kind() == reflect.Bool {
		return "false"
	}
	return ""
}

func trimSentence(value string) string {
	return strings.TrimSpace(value)
}
