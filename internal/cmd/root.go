package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/alecthomas/kong"

	"gobankcli/internal/config"
	"gobankcli/internal/outfmt"
)

type Globals struct {
	Config  string `help:"Config file path." type:"path"`
	DB      string `help:"SQLite archive path." type:"path"`
	JSON    bool   `help:"Emit stable JSON."`
	Plain   bool   `help:"Emit simple parseable plain text."`
	NoInput bool   `help:"Never prompt or wait for input."`
	Version bool   `help:"Print version and exit."`
}

type CLI struct {
	Globals

	Accounts     AccountsCmd     `cmd:"" help:"Fetch and archive accounts for a connection."`
	Connect      ConnectCmd      `cmd:"" help:"Start a read-only bank consent flow."`
	Doctor       DoctorCmd       `cmd:"" help:"Check local config, archive, and provider credentials."`
	Export       ExportCmd       `cmd:"" help:"Export normalized transactions as CSV."`
	Institutions InstitutionsCmd `cmd:"" help:"List provider institutions by country."`
	Init         InitCmd         `cmd:"" help:"Write a starter config and create local directories."`
	Status       StatusCmd       `cmd:"" help:"Show local archive status."`
	Sync         SyncCmd         `cmd:"" help:"Fetch and archive transactions for a connection."`
}

type App struct {
	Version    string
	Config     config.Config
	Out        *outfmt.Writer
	OutputMode outfmt.Mode
	Stdout     io.Writer
	Stderr     io.Writer
}

func Run(ctx context.Context, args []string, version string, stdout, stderr io.Writer) error {
	if hasVersionFlag(args) {
		fmt.Fprintln(stdout, version)
		return nil
	}

	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name("gobankcli"),
		kong.Description("Local-first read-only bank transaction archive."),
		kong.UsageOnError(),
	)
	if err != nil {
		return err
	}

	kctx, err := parser.Parse(args)
	if err != nil {
		return err
	}
	if cli.Version {
		fmt.Fprintln(stdout, version)
		return nil
	}
	if cli.JSON && cli.Plain {
		return errors.New("--json and --plain cannot be used together")
	}

	cfg, err := config.Load(cli.Config)
	if err != nil {
		if kctx.Command() != "init" || !cli.Init.Force {
			return err
		}
		cfg = config.Default()
		if cli.Config != "" {
			cfg.SourcePath = config.ExpandPath(cli.Config)
		}
		cfg.Expand()
	}
	if cli.DB != "" {
		cfg.Paths.DB = cli.DB
	}
	cfg.Expand()

	mode := outfmt.ModeHuman
	if cli.JSON {
		mode = outfmt.ModeJSON
	} else if cli.Plain {
		mode = outfmt.ModePlain
	}

	app := &App{
		Version:    version,
		Config:     cfg,
		Out:        outfmt.New(stdout, mode),
		OutputMode: mode,
		Stdout:     stdout,
		Stderr:     stderr,
	}
	kctx.Bind(app)
	kctx.BindTo(ctx, (*context.Context)(nil))
	return kctx.Run()
}

func hasVersionFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--version" {
			return true
		}
	}
	return false
}
