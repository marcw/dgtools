package discogs

import (
	"compress/gzip"
	"encoding/xml"
	"io"
	"os"
	"regexp"
	"strings"
)

var typeExtractor = regexp.MustCompile(`(artists|releases|masters|labels)`)
var dateExtractor = regexp.MustCompile(`discogs_(\d{4})(\d{2})`)

type DumpFilename string

func (fn DumpFilename) String() string {
	return string(fn)
}

func (fn DumpFilename) Gzipped() bool {
	return strings.HasSuffix(string(fn), ".gz")
}

func (fn DumpFilename) Year() string {
	return dateExtractor.FindStringSubmatch(string(fn))[1]
}

func (fn DumpFilename) Month() string {
	return dateExtractor.FindStringSubmatch(string(fn))[2]
}

func (fn DumpFilename) Type() string {
	return typeExtractor.FindStringSubmatch(string(fn))[1]
}

// Dump is a wrapper around a discogs dump file.
// Dump implements the io.ReadCloser interface.
type Dump struct {
	reader  io.Reader
	Decoder *xml.Decoder

	file     *os.File
	gzReader *gzip.Reader
}

// OpenDiscogsDump creates a new DiscogsDump.
func OpenDumpFile(filename string) (*Dump, error) {
	dd := &Dump{}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	dd.file = file
	dd.reader = file

	if !strings.HasSuffix(filename, ".gz") {
		return dd, nil
	}

	gz, err := gzip.NewReader(file)
	if err != nil {
		dd.file.Close()
		return nil, err
	}
	dd.gzReader = gz
	dd.reader = gz

	dd.Decoder = xml.NewDecoder(dd.reader)

	return dd, nil
}

// Close closes the dump file.
func (dd *Dump) Close() error {
	if dd.gzReader != nil {
		if err := dd.gzReader.Close(); err != nil {
			return err
		}
	}

	return dd.file.Close()
}

func (dd *Dump) Read(p []byte) (n int, err error) {
	return dd.reader.Read(p)
}
