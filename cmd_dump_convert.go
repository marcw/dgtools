package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/marcw/dgtools/internal/discogs"
	"github.com/parquet-go/parquet-go"
	"github.com/urfave/cli/v3"
)

const (
	FormatParquet = "parquet"
	FormatNdjson  = "ndjson"
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
			Usage: "Sets the output format for the conversion",
			Value: FormatParquet,
			Action: func(ctx context.Context, cmd *cli.Command, s string) error {
				if s != FormatParquet && s != FormatNdjson {
					return fmt.Errorf("supported formats are: %s, %s", FormatParquet, FormatNdjson)
				}
				return nil
			},
		},
		&cli.BoolFlag{
			Name:  "no-progress",
			Usage: "Don't show any progress bar",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "out",
			Usage: "Save the converted output to file",
		},
		&cli.Int64Flag{
			Name:  "stop-after",
			Usage: "Stop conversion after X records",
			Value: 0,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		outputFormat := cmd.String("format")
		outputFile := cmd.String("out")
		inputFile := cmd.StringArg("name")
		noProgress := cmd.Bool("no-progress")

		i := int64(0)
		var outFile *os.File
		var parquetWriter *parquet.Writer

		// First, we validate the arguments and flags.
		if inputFile == "" {
			return fmt.Errorf("input file is required")
		}

		// do not output binary things to stdout
		if outputFormat == FormatParquet && outputFile == "" {
			return fmt.Errorf("output file is required for parquet format conversion")
		}
		// if we convert to ndjson, we don't output progress
		if outputFormat == FormatNdjson && outputFile == "" {
			noProgress = true
		}

		dump, err := discogs.OpenDumpFile(inputFile)
		if err != nil {
			return err
		}
		defer dump.Close()

		if outputFile != "" {
			outFile, err := os.Create(outputFile)
			if err != nil {
				return err
			}
			defer outFile.Close()
		} else {
			outFile = os.Stdout
		}

		if outputFormat == FormatParquet {
			parquetWriter := parquet.NewWriter(outFile)
			defer parquetWriter.Close()
		}

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		if !noProgress {
			s.Suffix = " Converting..."
			s.Start()
		}
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

			switch outputFormat {
			case FormatParquet:
				if err := parquetWriter.Write(element); err != nil {
					return err
				}
			case FormatNdjson:
				b, err := json.Marshal(element)
				if err != nil {
					return err
				}
				if _, err := outFile.Write(b); err != nil {
					return err
				}
				if _, err := outFile.Write([]byte("\n")); err != nil {
					return err
				}
			}

			i++
			if cmd.Int64("stop-after") != 0 && i >= cmd.Int64("stop-after") {
				break
			}
			if !noProgress && i%1000 == 0 {
				s.Suffix = fmt.Sprintf(" Converting... %d", i)
			}
		}

		switch outputFormat {
		case FormatParquet:
			if err := parquetWriter.Flush(); err != nil {
				return err
			}
		case FormatNdjson:
			if _, err := outFile.Write([]byte("\n")); err != nil {
				return err
			}
			if outFile != os.Stdout {
				if err := outFile.Sync(); err != nil {
					return err
				}
			}
		}

		if !noProgress {
			s.Stop()
			fmt.Printf("Converted %d rows in %s.\n", i, time.Since(now))
		}

		return nil
	},
}
