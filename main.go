package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

const VERSION = "0.3.0"

func main() {
	cmd := &cli.Command{
		Name:    "dgtools",
		Usage:   "Work with Discogs data dump",
		Version: VERSION,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "discogs-bucket",
				Usage: "The URL of the Discogs data dumps",
				Value: "https://discogs-data-dumps.s3.us-west-2.amazonaws.com",
			},
		},
		Commands: []*cli.Command{
			dumpCmd,
			dbCmd,
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
