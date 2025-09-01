package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

var dbCmd = &cli.Command{
	Name:  "db",
	Usage: "Work with a database",
	Commands: []*cli.Command{
		dbPrepareCmd,
		dbImportCmd,
		dbNukeCmd,
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "database-url",
			Value:   fmt.Sprintf("postgres://%s@localhost:5432/dgtools", os.Getenv("USER")),
			Usage:   "The URL of the database to connect to",
			Sources: cli.EnvVars("DATABASE_URL"),
		},
	},
}
