package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v3"
)

//go:embed migrations/pg/*.sql
var embedMigrations embed.FS

var dbPrepareCmd = &cli.Command{
	Name:  "prepare",
	Usage: "Prepare the database for import",
	Action: func(ctx context.Context, cmd *cli.Command) error {

		db, err := sql.Open("pgx", cmd.String("database-url"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		goose.SetBaseFS(embedMigrations)

		if err := goose.SetDialect("postgres"); err != nil {
			return err
		}

		if err := goose.Up(db, "migrations/pg"); err != nil {
			return err
		}

		fmt.Println("Database prepared")

		return nil
	},
}
