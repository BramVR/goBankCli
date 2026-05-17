package cmd

import "context"

type StatusCmd struct{}

type statusReport struct {
	DBPath       string `json:"db_path"`
	ArchiveOpen  bool   `json:"archive_open"`
	Transactions int64  `json:"transactions"`
	Accounts     int64  `json:"accounts"`
	Message      string `json:"message,omitempty"`
}

func (c StatusCmd) Run(_ context.Context, app *App) error {
	return app.Out.Write(statusReport{
		DBPath:      app.Config.Paths.DB,
		ArchiveOpen: false,
		Message:     "archive schema not initialized yet",
	})
}
