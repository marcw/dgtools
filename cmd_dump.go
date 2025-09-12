package main

import "github.com/urfave/cli/v3"

var dumpCmd = &cli.Command{
	Name:  "dump",
	Usage: "Work with Discogs data dump",
	Commands: []*cli.Command{
		discogsDumpListCmd,
		discogsDumpStructureCmd,
		discogsDumpDownloadCmd,
		discogsDumpConvertCmd,
	},
}
