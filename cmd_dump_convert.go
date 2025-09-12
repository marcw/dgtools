package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/marcw/dgtools/internal/discogs"
	"github.com/parquet-go/parquet-go"
	"github.com/urfave/cli/v3"
)

var discogsDumpConvertCmd = &cli.Command{
	Name:  "convert",
	Usage: "Convert a Discogs data dump to a different format",
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:      "name",
			UsageText: "The file to convert",
		},
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "format",
			Usage: "The output format",
			Value: "parquet",
		},
		&cli.StringFlag{
			Name:  "out",
			Usage: "Force the download even if the file already exists",
		},
		&cli.Int64Flag{
			Name:  "stop-after",
			Usage: "The mode to convert",
			Value: 0,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		outputFile := cmd.String("out")
		outputFormat := cmd.String("format")
		if outputFormat != "parquet" {
			return fmt.Errorf("only the parquet format is supported right now.")
		}

		inputFile := cmd.StringArg("name")
		if inputFile == "" {
			return fmt.Errorf("input file is required")
		}

		dump, err := discogs.OpenDumpFile(inputFile)
		if err != nil {
			return err
		}
		defer dump.Close()

		if outputFile == "" {
			return fmt.Errorf("output file is required")
		}

		outFile, err := os.Create(cmd.String("out"))
		if err != nil {
			return err
		}
		defer outFile.Close()
		writer := parquet.NewWriter(outFile)
		defer writer.Close()

		i := int64(0)
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Converting..."
		s.Start()

		now := time.Now()

		for {
			element, err := dump.DecodeNextElement()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if element == nil {
				continue
			}

			if err := writer.Write(element); err != nil {
				return err
			}

			i++
			if cmd.Int64("stop-after") != 0 && i >= cmd.Int64("stop-after") {
				break
			}
			if i%1000 == 0 {
				s.Suffix = fmt.Sprintf(" Converting... %d", i)
			}
		}
		if err := writer.Flush(); err != nil {
			return err
		}
		s.Stop()
		fmt.Printf("Converted %d rows in %s.\n", i, time.Since(now))

		return nil
	},
}
