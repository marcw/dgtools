package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marcw/dgtools/internal/discogs"
	"github.com/urfave/cli/v3"
)

var dbImportCmd = &cli.Command{
	Name:  "import",
	Usage: "Import data from a dump file to the database",
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:      "file",
			UsageText: "The file to import the data from",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		pool, err := pgxpool.New(context.Background(), cmd.String("database-url"))
		if err != nil {
			log.Fatal(err)
		}
		defer pool.Close()

		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Release()

		file := cmd.StringArg("file")
		if file == "" {
			return fmt.Errorf("file is required")
		}

		dumpFile := discogs.DumpFilename(file)
		var modes []int
		if dumpFile.Type() == "artists" {
			modes = []int{discogs.ModeArtists, discogs.ModeArtistsAliases, discogs.ModeArtistsMemberships}
		} else if dumpFile.Type() == "labels" {
			modes = []int{discogs.ModeLabels}
		} else if dumpFile.Type() == "masters" {
			modes = []int{discogs.ModeMasters, discogs.ModeMastersArtists}
		} else if dumpFile.Type() == "releases" {
			modes = []int{discogs.ModeReleases, discogs.ModeReleasesArtists, discogs.ModeReleasesExtraArtists, discogs.ModeReleasesLabels}
		}

		for _, mode := range modes {
			table := discogs.Tables[mode]
			fmt.Println("Truncating table", table.Sanitize())
			conn.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s CASCADE", discogs.Tables[mode].Sanitize()))
		}

		now := time.Now()
		CopyDiscogsDumpSinglePass(pool, file, modes)

		fmt.Printf("Processed dump in %s.\n", time.Since(now))
		return nil
	},
}

func CopyDiscogsDump(conn *pgx.Conn, filename string, mode int) (n int64, err error) {
	fmt.Printf("Processing %s in mode %d.\n", filename, mode)
	now := time.Now()
	var ds *discogs.CopyFromDump

	if ds, err = discogs.NewCopyFromDump(filename, mode); err != nil {
		return 0, err
	}
	defer ds.Close()

	n, err = conn.CopyFrom(context.Background(), ds.Table(), ds.Columns(), ds)
	duration := time.Since(now)
	log.Printf("Processed %d rows in %s.\n", n, duration)
	if err != nil {
		log.Println(err)
		return n, err
	}

	return
}

func CopyDiscogsDumpSinglePass(pool *pgxpool.Pool, filename string, modes []int) error {
	log.Printf("Processing %s in single-pass mode with %d tables.\n", filename, len(modes))
	now := time.Now()

	channelMap := make(map[int]chan []any)
	bufferSize := 1000

	for _, mode := range modes {
		channelMap[mode] = make(chan []any, bufferSize)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	parser, err := discogs.NewMultiTableXMLParser(filename, channelMap, &wg)
	if err != nil {
		return err
	}
	defer parser.Close()

	for _, mode := range modes {
		wg.Add(1)
		go func(mode int) {
			defer wg.Done()

			conn, err := pool.Acquire(context.Background())
			if err != nil {
				log.Printf("Failed to acquire connection for mode %d: %v", mode, err)
				return
			}
			defer conn.Release()

			source := discogs.NewCopyFromRecordChannel(channelMap[mode])
			n, err := conn.CopyFrom(context.Background(), discogs.Tables[mode], getColumnsForMode(mode), source)
			if err != nil {
				log.Printf("CopyFrom failed for mode %d: %v", mode, err)
				return
			}
			log.Printf("Mode %d: processed %d rows", mode, n)
		}(mode)
	}

	go func() {
		if err := parser.ParseAndDistribute(); err != nil {
			log.Printf("Parser error: %v", err)
		}
	}()

	wg.Wait()
	duration := time.Since(now)
	log.Printf("Single-pass processing completed in %s.\n", duration)
	return nil
}

func getColumnsForMode(mode int) []string {
	switch mode {
	case discogs.ModeArtists:
		return discogs.Artist{}.Columns()
	case discogs.ModeArtistsAliases:
		return discogs.Artist{}.AliasesColumns()
	case discogs.ModeArtistsMemberships:
		return discogs.Artist{}.MembershipsColumns()
	case discogs.ModeLabels:
		return discogs.Label{}.Columns()
	case discogs.ModeMasters:
		return discogs.Master{}.Columns()
	case discogs.ModeMastersArtists:
		return discogs.Master{}.ArtistsColumns()
	case discogs.ModeReleases:
		return discogs.Release{}.Columns()
	case discogs.ModeReleasesArtists:
		return discogs.Release{}.ArtistsColumns()
	case discogs.ModeReleasesExtraArtists:
		return discogs.Release{}.ExtraArtistsColumns()
	case discogs.ModeReleasesLabels:
		return discogs.Release{}.LabelsColumns()
	default:
		return nil
	}
}
