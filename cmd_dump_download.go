package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/urfave/cli/v3"
)

var discogsDumpDownloadCmd = &cli.Command{
	Name:  "download",
	Usage: "Download a Discogs data dump",
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:      "name",
			UsageText: "The file to download",
		},
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "out-dir",
			Usage: "The output directory",
			Value: ".",
		},
		&cli.BoolFlag{
			Name:  "overwrite",
			Usage: "Force the download even if the file already exists",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "checksum",
			Usage: "Check the checksum of the file after downloading",
			Value: true,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		name := cmd.StringArg("name")
		outDir := cmd.String("out-dir")
		overwrite := cmd.Bool("overwrite")
		checksum := cmd.Bool("checksum")

		if name == "" {
			return fmt.Errorf("name is required")
		}

		_, err := os.Stat(outDir)
		if os.IsNotExist(err) {
			return fmt.Errorf("output directory does not exist: %s", outDir)
		}

		if outDir == "." {
			outDir = ""
		}
		outFilename := filepath.Join(outDir, filepath.Base(name))
		if _, err := os.Stat(outFilename); !os.IsNotExist(err) && !overwrite {
			return fmt.Errorf("file already exists: %s", outFilename)
		}

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Downloading..."
		s.Start()
		url := fmt.Sprintf("%s/%s", cmd.String("discogs-bucket"), name)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		out := filepath.Base(name)
		outFile, err := os.Create(out)
		if err != nil {
			return err
		}
		defer outFile.Close()

		io.Copy(outFile, resp.Body)
		s.Stop()
		fmt.Println("Downloaded")

		if checksum {
			s.Suffix = " Verifying checksum..."
			s.Start()
			checksumFile := regexp.MustCompile("(artists|releases|masters|labels)").ReplaceAllString(name, "CHECKSUM")
			checksumFile = strings.Replace(checksumFile, ".xml.gz", ".txt", 1)
			resp, err := http.Get(fmt.Sprintf("%s/%s", cmd.String("discogs-bucket"), checksumFile))
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			outChecksumFile, err := os.Create(filepath.Join(outDir, filepath.Base(checksumFile)))
			if err != nil {
				return err
			}
			defer outChecksumFile.Close()
			io.Copy(outChecksumFile, resp.Body)

			// run sha256sum -c outChecksumFile in the output directory
			sha256sum := exec.Command("sha256sum", "--ignore-missing", "-c", filepath.Base(checksumFile))
			if outDir != "" {
				sha256sum.Dir = outDir
			}
			err = sha256sum.Run()
			if err != nil {
				return fmt.Errorf("failed to check checksum: %w", err)
			}
			s.Stop()
			fmt.Println("Checksum OK")
		}

		return nil
	},
}
