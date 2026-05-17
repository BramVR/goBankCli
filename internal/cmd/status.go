package cmd

import (
	"context"

	"gobankcli/internal/store"
)

type StatusCmd struct{}

type statusReport struct {
	DBPath       string `json:"db_path"`
	ArchiveOpen  bool   `json:"archive_open"`
	Institutions int64  `json:"institutions"`
	Connections  int64  `json:"connections"`
	Transactions int64  `json:"transactions"`
	Accounts     int64  `json:"accounts"`
	SyncRuns     int64  `json:"sync_runs"`
}

func (c StatusCmd) Run(ctx context.Context, app *App) error {
	s, err := store.Open(ctx, app.Config.Paths.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	status, err := s.Status(ctx)
	if err != nil {
		return err
	}
	return app.Out.Write(statusReport{
		DBPath:       status.DBPath,
		ArchiveOpen:  true,
		Institutions: status.Institutions,
		Connections:  status.Connections,
		Transactions: status.Transactions,
		Accounts:     status.Accounts,
		SyncRuns:     status.SyncRuns,
	})
}
