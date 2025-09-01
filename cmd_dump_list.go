package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/table"
	"github.com/marcw/dgtools/internal/discogs"
	"github.com/urfave/cli/v3"
)

type dump struct {
	Name         string
	Size         int64
	Type         string
	Year         string
	Month        string
	LastModified time.Time
}

var discogsDumpListCmd = &cli.Command{
	Name:  "list",
	Usage: "List the files in the Discogs data dumps",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "year",
			Usage: "Filter by year",
		},
		&cli.StringFlag{
			Name:  "month",
			Usage: "Filter by month",
		},
		&cli.StringFlag{
			Name:  "type",
			Usage: "Filter by data type",
		},
		&cli.BoolFlag{
			Name:  "no-table",
			Usage: "Don't print the table",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		discogsDataDumpsURL := cmd.String("discogs-bucket")

		resp, err := http.Get(discogsDataDumpsURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		decoder := xml.NewDecoder(resp.Body)
		var result listBucketResult
		if err := decoder.Decode(&result); err != nil {
			return err
		}

		dumps := make([]dump, 0)

		for _, content := range result.Contents {
			if !strings.HasSuffix(content.Key, ".xml.gz") {
				continue
			}

			dumpFilename := discogs.DumpFilename(content.Key)

			if cmd.String("year") != "" && dumpFilename.Year() != cmd.String("year") {
				continue
			}
			if cmd.String("month") != "" && dumpFilename.Month() != cmd.String("month") {
				continue
			}
			if cmd.String("type") != "" && dumpFilename.Type() != cmd.String("type") {
				continue
			}

			d := dump{
				Name:         content.Key,
				Size:         content.Size,
				LastModified: content.LastModified,
				Type:         dumpFilename.Type(),
				Year:         dumpFilename.Year(),
				Month:        dumpFilename.Month(),
			}

			dumps = append(dumps, d)
		}

		if !cmd.Bool("no-table") {
			t := table.New().Headers("YEAR", "MONTH", "TYPE", "NAME", "SIZE")
			for _, dump := range dumps {
				t.Row(dump.Year, dump.Month, dump.Type, dump.Name, fmt.Sprintf("%d", dump.Size))
			}
			fmt.Println(t)
		} else {
			for _, dump := range dumps {
				fmt.Println(dump.Name)
			}
		}

		return nil
	},
}

type listBucketResult struct {
	Name     string                     `xml:"Name"`
	Contents []listBucketResultContents `xml:"Contents"`
}

type listBucketResultContents struct {
	Key          string    `xml:"Key"`
	Size         int64     `xml:"Size"`
	LastModified time.Time `xml:"LastModified"`
}
