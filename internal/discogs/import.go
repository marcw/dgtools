package discogs

import (
	"encoding/xml"
	"fmt"
	"io"
	"slices"
	"sync"

	"github.com/jackc/pgx/v5"
)

const (
	ModeArtists = iota
	ModeArtistsAliases
	ModeArtistsMemberships
	ModeLabels
	ModeMasters
	ModeMastersArtists
	ModeReleases
	ModeReleasesArtists
	ModeReleasesExtraArtists
	ModeReleasesLabels
)

var Modes = []int{
	ModeArtists,
	ModeArtistsAliases,
	ModeArtistsMemberships,
	ModeLabels,
	ModeMasters,
	ModeMastersArtists,
	ModeReleases,
	ModeReleasesArtists,
	ModeReleasesExtraArtists,
	ModeReleasesLabels,
}

var Tables = map[int]pgx.Identifier{
	ModeArtists:              pgx.Identifier{"discogs_artists"},
	ModeArtistsAliases:       pgx.Identifier{"discogs_artists_aliases"},
	ModeArtistsMemberships:   pgx.Identifier{"discogs_artists_members"},
	ModeLabels:               pgx.Identifier{"discogs_labels"},
	ModeMasters:              pgx.Identifier{"discogs_masters"},
	ModeMastersArtists:       pgx.Identifier{"discogs_master_artists"},
	ModeReleases:             pgx.Identifier{"discogs_releases"},
	ModeReleasesArtists:      pgx.Identifier{"discogs_release_artists"},
	ModeReleasesExtraArtists: pgx.Identifier{"discogs_release_extra_artists"},
	ModeReleasesLabels:       pgx.Identifier{"discogs_release_labels"},
}

type CopyFromDump struct {
	mode    int
	dd      *Dump
	decoder *xml.Decoder
	records [][]any
	err     error
}

// NewCopyFromDump opens a Discogs dump file and returns a CopyFromDump.
func NewCopyFromDump(filename string, mode int) (ds *CopyFromDump, err error) {
	if !slices.Contains(Modes, mode) {
		return nil, fmt.Errorf("cannot instantiate a CopyFromDump with unknown mode: %d", mode)
	}

	ds = &CopyFromDump{
		mode: mode,
	}

	dd, err := OpenDumpFile(filename)
	if err != nil {
		return nil, err
	}

	ds.dd = dd
	ds.decoder = xml.NewDecoder(dd)

	return ds, nil
}

// Table returns the name of the table the data should be copied to.
func (ds *CopyFromDump) Table() pgx.Identifier {
	return Tables[ds.mode]
}

func (ds *CopyFromDump) Columns() []string {
	switch ds.mode {
	case ModeArtists:
		return Artist{}.Columns()
	case ModeArtistsAliases:
		return Artist{}.AliasesColumns()
	case ModeArtistsMemberships:
		return Artist{}.MembershipsColumns()
	case ModeLabels:
		return Label{}.Columns()
	case ModeMasters:
		return Master{}.Columns()
	case ModeMastersArtists:
		return Master{}.ArtistsColumns()
	case ModeReleases:
		return Release{}.Columns()
	case ModeReleasesArtists:
		return Release{}.ArtistsColumns()
	case ModeReleasesExtraArtists:
		return Release{}.ExtraArtistsColumns()
	case ModeReleasesLabels:
		return Release{}.LabelsColumns()
	default:
		return nil
	}
}

func (ds *CopyFromDump) Close() error {
	return ds.dd.Close()
}

// See pgx.CopyFromSource interface for more details.
func (ds *CopyFromDump) Err() error {
	return ds.err
}

// Next returns true if there is another row and makes the next row data
// available to Values(). When there are no more rows available or an error
// has occurred it returns false.
// See pgx.CopyFromSource interface for more details.
func (ds *CopyFromDump) Next() bool {
	// If there are records left to process, return true
	// Values() will return the next record automatically.
	if len(ds.records) > 0 {
		return true
	}

	// If there are no records left to process, decode the next XML element
	// and replace the records slice with the new record(s).
	decodedElement, err := ds.decodeXML()
	if decodedElement == nil && err == nil {
		ds.err = nil
		ds.records = [][]any{}

		return false
	}
	if err != nil {
		ds.err = err
		return false
	}
	ds.err = nil

	// TODO: handle the decodedElement
	switch ds.mode {
	case ModeArtists:
		ds.records = [][]any{decodedElement.(*Artist).ToRecord()}
	case ModeArtistsAliases:
		ds.records = decodedElement.(*Artist).ToAliasesRecords()
	case ModeArtistsMemberships:
		ds.records = decodedElement.(*Artist).ToMembershipsRecords()
	case ModeLabels:
		ds.records = [][]any{decodedElement.(*Label).ToRecord()}
	case ModeMasters:
		ds.records = [][]any{decodedElement.(*Master).ToRecord()}
	case ModeMastersArtists:
		ds.records = decodedElement.(*Master).ToArtistsRecords()
	case ModeReleases:
		ds.records = [][]any{decodedElement.(*Release).ToRecord()}
	case ModeReleasesArtists:
		ds.records = decodedElement.(*Release).ToArtistsRecords()
	case ModeReleasesExtraArtists:
		ds.records = decodedElement.(*Release).ToExtraArtistsRecords()
	case ModeReleasesLabels:
		ds.records = decodedElement.(*Release).ToLabelsRecords()
	}

	// If there are no
	if len(ds.records) == 0 {
		return ds.Next()
	}

	return true
}

// Values returns the values for the current row.
// See pgx.CopyFromSource interface for more details.
func (ds *CopyFromDump) Values() (value []any, err error) {
	value, ds.records = ds.records[0], ds.records[1:]

	return
}

func (ds *CopyFromDump) decodeXML() (any, error) {
	for {
		t, err := ds.decoder.Token()
		if err == io.EOF {
			return nil, nil
		}

		if err != nil {
			return nil, err
		}

		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			inElement := se.Name.Local
			if inElement == "artist" {
				artist := &Artist{}
				ds.decoder.DecodeElement(artist, &se)
				return artist, nil
			}
			if inElement == "label" {
				label := &Label{}
				ds.decoder.DecodeElement(label, &se)
				return label, nil
			}
			if inElement == "master" {
				master := &Master{}
				ds.decoder.DecodeElement(master, &se)
				return master, nil
			}
			if inElement == "release" {
				release := &Release{}
				ds.decoder.DecodeElement(release, &se)
				return release, nil
			}
		default:
		}
	}
}

// CopyFromRecordChannel implements pgx.CopyFromSource for channel-based data
type CopyFromRecordChannel struct {
	recordChan <-chan []any
	currentRow []any
	err        error
	done       bool
}

func NewCopyFromRecordChannel(recordChan <-chan []any) *CopyFromRecordChannel {
	return &CopyFromRecordChannel{
		recordChan: recordChan,
	}
}

func (c *CopyFromRecordChannel) Next() bool {
	if c.done {
		return false
	}

	row, ok := <-c.recordChan
	if !ok {
		c.done = true
		return false
	}

	c.currentRow = row
	return true
}

func (c *CopyFromRecordChannel) Values() ([]any, error) {
	return c.currentRow, c.err
}

func (c *CopyFromRecordChannel) Err() error {
	return c.err
}

// MultiTableXMLParser parses XML once and distributes records to multiple channels
type MultiTableXMLParser struct {
	channels map[int]chan []any
	wg       *sync.WaitGroup
	dd       *Dump
}

func NewMultiTableXMLParser(filename string, channelMap map[int]chan []any, wg *sync.WaitGroup) (*MultiTableXMLParser, error) {
	parser := &MultiTableXMLParser{
		channels: channelMap,
		wg:       wg,
	}

	dd, err := OpenDumpFile(filename)
	if err != nil {
		return nil, err
	}

	parser.dd = dd
	return parser, nil
}

func (p *MultiTableXMLParser) Close() error {
	return p.dd.Close()
}

func (p *MultiTableXMLParser) ParseAndDistribute() error {
	defer func() {
		for _, ch := range p.channels {
			close(ch)
		}
		p.wg.Done()
	}()

	for {
		t, err := p.dd.Decoder.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch se := t.(type) {
		case xml.StartElement:
			inElement := se.Name.Local
			switch inElement {
			case "artist":
				artist := &Artist{}
				if err := p.dd.Decoder.DecodeElement(artist, &se); err != nil {
					return err
				}
				p.distributeArtistRecords(artist)
			case "label":
				label := &Label{}
				if err := p.dd.Decoder.DecodeElement(label, &se); err != nil {
					return err
				}
				p.distributeLabelRecords(label)
			case "master":
				master := &Master{}
				if err := p.dd.Decoder.DecodeElement(master, &se); err != nil {
					return err
				}
				p.distributeMasterRecords(master)
			case "release":
				release := &Release{}
				if err := p.dd.Decoder.DecodeElement(release, &se); err != nil {
					return err
				}
				p.distributeReleaseRecords(release)
			}
		}
	}
}

func (p *MultiTableXMLParser) distributeArtistRecords(artist *Artist) {
	if ch, exists := p.channels[ModeArtists]; exists {
		ch <- artist.ToRecord()
	}
	if ch, exists := p.channels[ModeArtistsAliases]; exists {
		for _, record := range artist.ToAliasesRecords() {
			ch <- record
		}
	}
	if ch, exists := p.channels[ModeArtistsMemberships]; exists {
		for _, record := range artist.ToMembershipsRecords() {
			ch <- record
		}
	}
}

func (p *MultiTableXMLParser) distributeLabelRecords(label *Label) {
	if ch, exists := p.channels[ModeLabels]; exists {
		ch <- label.ToRecord()
	}
}

func (p *MultiTableXMLParser) distributeMasterRecords(master *Master) {
	if ch, exists := p.channels[ModeMasters]; exists {
		ch <- master.ToRecord()
	}
	if ch, exists := p.channels[ModeMastersArtists]; exists {
		for _, record := range master.ToArtistsRecords() {
			ch <- record
		}
	}
}

func (p *MultiTableXMLParser) distributeReleaseRecords(release *Release) {
	if ch, exists := p.channels[ModeReleases]; exists {
		ch <- release.ToRecord()
	}
	if ch, exists := p.channels[ModeReleasesArtists]; exists {
		for _, record := range release.ToArtistsRecords() {
			ch <- record
		}
	}
	if ch, exists := p.channels[ModeReleasesExtraArtists]; exists {
		for _, record := range release.ToExtraArtistsRecords() {
			ch <- record
		}
	}
	if ch, exists := p.channels[ModeReleasesLabels]; exists {
		for _, record := range release.ToLabelsRecords() {
			ch <- record
		}
	}
}
