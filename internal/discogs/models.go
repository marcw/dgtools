package discogs

import (
	"encoding/xml"
)

type Artist struct {
	ID          int64   `xml:"id" parquet:"id,zstd"`
	Name        string  `xml:"name" parquet:"name,zstd"`
	RealName    *string `xml:"realname" parquet:"real_name,zstd"`
	Profile     *string `xml:"profile" parquet:"profile,zstd"`
	DataQuality string  `xml:"data_quality" parquet:"data_quality,dict"`

	URLs           []string `xml:"urls>url" parquet:"urls,zstd"`
	Aliases        []*Name  `xml:"aliases>name" parquet:"aliases"`
	NameVariations []string `xml:"namevariations>name" parquet:"name_variations,zstd"`
	Members        []*Name  `xml:"members>name" parquet:"members"`
	Groups         []*Name  `xml:"groups>name" parquet:"groups"`
}

type Name struct {
	ID   int64  `xml:"id,attr" parquet:"id,zstd"`
	Name string `xml:",chardata" parquet:"name,zstd"`
}

func (a Artist) Columns() []string {
	return []string{
		"id",
		"name",
		"real_name",
		"profile",
		"data_quality",
		"name_variations",
		"urls",
	}
}

func (a Artist) AliasesColumns() []string {
	return []string{"artist_id", "alias_id"}
}

func (a Artist) MembershipsColumns() []string {
	return []string{"artist_id", "member_id"}
}

func (a *Artist) clean() {
	if a.RealName != nil && *a.RealName == "" {
		a.RealName = nil
	}
	if a.Profile != nil && *a.Profile == "" {
		a.Profile = nil
	}
}

func (a *Artist) ToRecord() []any {
	return []any{
		a.ID,
		a.Name,
		a.RealName,
		a.Profile,
		a.DataQuality,
		a.NameVariations,
		a.URLs,
	}
}

func (a *Artist) ToAliasesRecords() [][]any {
	records := make([][]any, len(a.Aliases))
	for i := range a.Aliases {
		records[i] = []any{a.Aliases[i].ID, a.Aliases[i].ID}
	}
	return records
}

func (a *Artist) ToMembershipsRecords() [][]any {
	records := make([][]any, len(a.Members))
	for i := range a.Members {
		records[i] = []any{a.Members[i].ID, a.Members[i].ID}
	}
	return records
}

type label struct {
	ID          int64       `xml:"id" parquet:"id,zstd"`
	Name        string      `xml:"name" parquet:"name,zstd"`
	ContactInfo *string     `xml:"contactinfo" parquet:"contact_info,zstd"`
	Profile     *string     `xml:"profile" parquet:"profile,zstd"`
	DataQuality string      `xml:"data_quality" parquet:"data_quality,dict"`
	URLs        []string    `xml:"urls>url" parquet:"urls,zstd"`
	SubLabels   []*SubLabel `xml:"sublabels>label" parquet:"sub_labels"`
}

type Label struct {
	label
	ParentLabelID *int64 `parquet:"parent_label_id,zstd"`
}

func (l *Label) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var tmpLabel struct {
		label
		ParentLabel SubLabel `xml:"parentLabel"`
	}

	if err := d.DecodeElement(&tmpLabel, &start); err != nil {
		return err
	}
	l.label = tmpLabel.label
	if tmpLabel.ParentLabel.ID != 0 {
		l.ParentLabelID = &tmpLabel.ParentLabel.ID
	}
	l.clean()

	return nil
}

func (l *Label) clean() {
	if l.ContactInfo != nil && *l.ContactInfo == "" {
		l.ContactInfo = nil
	}
	if l.Profile != nil && *l.Profile == "" {
		l.Profile = nil
	}
}

func (l Label) Columns() []string {
	return []string{
		"id",
		"parent_label_id",
		"data_quality",
		"name",
		"profile",
		"contact_info",
		"urls",
	}
}

type SubLabel struct {
	ID   int64  `xml:"id,attr" parquet:"id,zstd"`
	Name string `xml:",chardata" parquet:"name,zstd"`
}

func (l *Label) ToRecord() []any {
	return []any{
		l.ID,
		l.ParentLabelID,
		l.DataQuality,
		l.Name,
		l.Profile,
		l.ContactInfo,
		l.URLs,
	}
}

type master struct {
	ID            int64           `xml:"id,attr" parquet:"id,zstd"`
	Artists       []*MasterArtist `xml:"artists>artist" parquet:"artists"`
	MainReleaseID *int64          `xml:"main_release" parquet:"main_release_id,zstd"`
	DataQuality   string          `xml:"data_quality" parquet:"data_quality,dict"`
	Videos        []Video         `xml:"videos>video" parquet:"videos"`
	Title         string          `xml:"title" parquet:"title,zstd"`
	Year          *int32          `xml:"year" parquet:"year,dict"`
	Genres        []string        `xml:"genres>genre" parquet:"genres,dict"`
	Styles        []string        `xml:"styles>style" parquet:"styles,dict"`
	Notes         *string         `xml:"notes" parquet:"notes,zstd"`
}

type Master struct {
	master
}

func (m *Master) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var temp struct {
		master
	}

	if err := d.DecodeElement(&temp, &start); err != nil {
		return err
	}

	m.master = temp.master
	m.clean()
	return nil
}

func (m *Master) clean() {
	if m.Year != nil && *m.Year == 0 {
		m.Year = nil
	}

	if m.MainReleaseID != nil && *m.MainReleaseID == 0 {
		m.MainReleaseID = nil
	}

	if m.Notes != nil && *m.Notes == "" {
		m.Notes = nil
	}

	for i := range m.Artists {
		m.Artists[i].clean()
	}
}

func (m Master) Columns() []string {
	return []string{
		"id",
		"main_release_id",
		"data_quality",
		"title",
		"year",
		"genres",
		"styles",
		"videos",
	}
}

func (m Master) ArtistsColumns() []string {
	return []string{"master_id", "artist_id", "name", "name_variation", "join"}
}

func (m *Master) ToRecord() []any {
	return []any{
		m.ID,
		m.MainReleaseID,
		m.DataQuality,
		m.Title,
		m.Year,
		m.Genres,
		m.Styles,
		m.Videos,
	}
}

func (m *Master) ToArtistsRecords() [][]any {
	artists := make([][]any, len(m.Artists))
	for i := range m.Artists {
		artists[i] = []any{
			m.ID,
			m.Artists[i].ID,
			m.Artists[i].Name,
			m.Artists[i].Anv,
			m.Artists[i].Join,
		}
	}

	return artists
}

type MasterArtist struct {
	ID   int64   `xml:"id" parquet:"id,zstd"`
	Name string  `xml:"name" parquet:"name,zstd"`
	Anv  *string `xml:"anv" parquet:"name_variation,zstd"`
	Join *string `xml:"join" parquet:"join,dict"`
}

func (m *MasterArtist) clean() {
	if m.Anv != nil && *m.Anv == "" {
		m.Anv = nil
	}
	if m.Join != nil && *m.Join == "" {
		m.Join = nil
	}
}

type Video struct {
	Src         string `xml:"src,attr" json:"src" parquet:"src,zstd"`
	Duration    int32  `xml:"duration,attr" json:"duration" parquet:"duration,zstd"`
	Embed       string `xml:"embed,attr" json:"embed" parquet:"embed,dict"`
	Title       string `xml:"title" json:"title" parquet:"title,zstd"`
	Description string `xml:"description" json:"description" parquet:"description,zstd"`
}

type release struct {
	ID          int64   `xml:"id,attr" parquet:"id"`
	Status      string  `xml:"status,attr" parquet:"status,dict"`
	Country     *string `xml:"country" parquet:"country,dict"`
	Released    *string `xml:"released" parquet:"released,dict"`
	Notes       *string `xml:"notes" parquet:"notes,zstd"`
	DataQuality string  `xml:"data_quality" parquet:"data_quality,dict"`
	Title       string  `xml:"title" parquet:"title,zstd"`

	Artists      []*MasterArtist  `xml:"artists>artist" parquet:"artists"`
	Companies    []*Company       `xml:"companies>company" parquet:"companies"`
	ExtraArtists []*ExtraArtist   `xml:"extraartists>artist" parquet:"extra_artists"`
	Formats      []*ReleaseFormat `xml:"formats>format" parquet:"formats"`
	Genres       []string         `xml:"genres>genre" parquet:"genres,dict"`
	Identifiers  []*Identifier    `xml:"identifiers>identifier" parquet:"identifiers"`
	Labels       []*ReleaseLabel  `xml:"labels>label" parquet:"labels"`
	Series       []*Serie         `xml:"series>serie" parquet:"series"`
	Styles       []string         `xml:"styles>style" parquet:"styles,dict"`
	Tracklist    []*Track         `xml:"tracklist>track" parquet:"tracklist"`
	Videos       []*Video         `xml:"videos>video" parquet:"videos"`
}

type Release struct {
	release
	MasterID      *int64 `parquet:"master_id"`
	IsMainRelease bool   `parquet:"is_main_release,zstd"`
}

func (r *Release) clean() {
	if r.Notes != nil && *r.Notes == "" {
		r.Notes = nil
	}

	if r.Released != nil && *r.Released == "" {
		r.Released = nil
	}

	if r.Country != nil && *r.Country == "" {
		r.Country = nil
	}

	for i := range r.Artists {
		r.Artists[i].clean()
	}
	for i := range r.Companies {
		r.Companies[i].clean()
	}
	for i := range r.ExtraArtists {
		r.ExtraArtists[i].clean()
	}
	for i := range r.Identifiers {
		r.Identifiers[i].clean()
	}
	for i := range r.Labels {
		r.Labels[i].clean()
	}
	for i := range r.Series {
		r.Series[i].clean()
	}
	for i := range r.Tracklist {
		r.Tracklist[i].clean()
	}
}

func (r *Release) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var temp struct {
		release
		MasterID masterID `xml:"master_id"`
	}

	if err := d.DecodeElement(&temp, &start); err != nil {
		return err
	}

	r.release = temp.release
	if temp.MasterID.MasterID != 0 {
		r.MasterID = &temp.MasterID.MasterID
	}
	r.IsMainRelease = temp.MasterID.IsMainRelease
	r.clean()

	return nil
}

func (r Release) Columns() []string {
	return []string{
		"id",
		"master_id",
		"is_main_release",
		"status",
		"title",
		"country",
		"released",
		"notes",
		"data_quality",
		"genres",
		"styles",
		"videos",
		"formats",
		"tracklist",
		"companies",
		"identifiers",
		"series",
	}
}

func (r Release) ArtistsColumns() []string {
	return []string{"release_id", "artist_id", "name", "name_variation", "join"}
}

func (r Release) ExtraArtistsColumns() []string {
	return []string{"release_id", "artist_id", "name", "name_variation", "role"}
}

func (r Release) LabelsColumns() []string {
	return []string{"release_id", "label_id", "name", "catno"}
}

func (r *Release) ToArtistsRecords() [][]any {
	artists := make([][]any, len(r.Artists))
	for i := range r.Artists {
		artists[i] = []any{
			r.ID,
			r.Artists[i].ID,
			r.Artists[i].Name,
			r.Artists[i].Anv,
			r.Artists[i].Join,
		}
	}

	return artists
}

func (r *Release) ToRecord() []any {
	return []any{
		r.ID,
		r.MasterID,
		r.IsMainRelease,
		r.Status,
		r.Title,
		r.Country,
		r.Released,
		r.Notes,
		r.DataQuality,
		r.Genres,
		r.Styles,
		r.Videos,
		r.Formats,
		r.Tracklist,
		r.Companies,
		r.Identifiers,
		r.Series,
	}
}

func (r *Release) ToExtraArtistsRecords() [][]any {
	extraArtists := make([][]any, len(r.ExtraArtists))
	for i := range r.ExtraArtists {
		extraArtists[i] = []any{
			r.ID,
			r.ExtraArtists[i].ID,
			r.ExtraArtists[i].Name,
			r.ExtraArtists[i].Anv,
			r.ExtraArtists[i].Role,
		}
	}

	return extraArtists
}

func (r *Release) ToLabelsRecords() [][]any {
	labels := make([][]any, len(r.Labels))
	for i := range r.Labels {
		labels[i] = []any{
			r.ID,
			r.Labels[i].ID,
			r.Labels[i].Name,
			r.Labels[i].Catno,
		}
	}

	return labels
}

type masterID struct {
	MasterID      int64 `xml:",chardata"`
	IsMainRelease bool  `xml:"is_main_release,attr"`
}

type SubTrack struct {
	Position     *string         `xml:"position" parquet:"position,dict"`
	Title        string          `xml:"title" parquet:"title,zstd"`
	Duration     *string         `xml:"duration" parquet:"duration,zstd"`
	Artists      []*MasterArtist `xml:"artists>artist" parquet:"artists"`
	ExtraArtists []*ExtraArtist  `xml:"extraartists>artist" parquet:"extra_artists"`
}

func (t *SubTrack) clean() {
	if t.Position != nil && *t.Position == "" {
		t.Position = nil
	}
	if t.Duration != nil && *t.Duration == "" {
		t.Duration = nil
	}
	for i := range t.Artists {
		t.Artists[i].clean()
	}
	for i := range t.ExtraArtists {
		t.ExtraArtists[i].clean()
	}
}

type Track struct {
	Position     *string         `xml:"position" parquet:"position,dict"`
	Title        string          `xml:"title" parquet:"title,zstd"`
	Duration     *string         `xml:"duration" parquet:"duration,zstd"`
	Artists      []*MasterArtist `xml:"artists>artist" parquet:"artists"`
	ExtraArtists []*ExtraArtist  `xml:"extraartists>artist" parquet:"extra_artists"`
	SubTracks    []*SubTrack     `xml:"sub_tracks>track" parquet:"sub_tracks"`
}

func (t *Track) clean() {
	if t.Position != nil && *t.Position == "" {
		t.Position = nil
	}
	if t.Duration != nil && *t.Duration == "" {
		t.Duration = nil
	}

	for i := range t.Artists {
		t.Artists[i].clean()
	}
	for i := range t.ExtraArtists {
		t.ExtraArtists[i].clean()
	}
	for i := range t.SubTracks {
		t.SubTracks[i].clean()
	}
}

type Identifier struct {
	Type        string  `xml:"type,attr" parquet:"type,dict"`
	Description *string `xml:"description,attr" parquet:"description,zstd"`
	Value       string  `xml:"value,attr" parquet:"value,zstd"`
}

func (i *Identifier) clean() {
	if i.Description != nil && *i.Description == "" {
		i.Description = nil
	}
}

type ExtraArtist struct {
	ID   int64   `xml:"id" parquet:"id,zstd"`
	Name string  `xml:"name" parquet:"name,zstd"`
	Anv  *string `xml:"anv" parquet:"name_variation,zstd"`
	Role *string `xml:"role" parquet:"role,dict"`
}

func (e *ExtraArtist) clean() {
	if e.Anv != nil && *e.Anv == "" {
		e.Anv = nil
	}

	if e.Role != nil && *e.Role == "" {
		e.Role = nil
	}
}

type ReleaseLabel struct {
	ID    int64   `xml:"id,attr" parquet:"id,zstd"`
	Name  string  `xml:"name,attr" parquet:"name,zstd"`
	Catno *string `xml:"catno,attr" parquet:"catno,zstd"`
}

func (r *ReleaseLabel) clean() {
	if r.Catno != nil && *r.Catno == "" {
		r.Catno = nil
	}
}

type Company struct {
	ID             int64   `xml:"id" parquet:"id,zstd"`
	Name           string  `xml:"name" parquet:"name,zstd"`
	EntityType     int64   `xml:"entity_type" parquet:"entity_type,zstd"`
	EntityTypeName string  `xml:"entity_type_name" parquet:"entity_type_name,zstd"`
	ResourceURL    string  `xml:"resource_url" parquet:"resource_url,zstd"`
	Catno          *string `xml:"catno" parquet:"catno,zstd"`
}

func (c *Company) clean() {
	if c.Catno != nil && *c.Catno == "" {
		c.Catno = nil
	}
}

type ReleaseFormat struct {
	Name         string   `xml:"name,attr" parquet:"name,zstd"`
	Qty          string   `xml:"qty,attr" parquet:"qty,zstd"`
	Text         string   `xml:"text,attr" parquet:"text,zstd"`
	Descriptions []string `xml:"descriptions>description" parquet:"descriptions,zstd"`
}

type Serie struct {
	ID    int64   `xml:"id,attr" parquet:"id,zstd"`
	Name  string  `xml:"name,attr" parquet:"name,zstd"`
	Catno *string `xml:"catno,attr" parquet:"catno,zstd"`
}

func (s *Serie) clean() {
	if s.Catno != nil && *s.Catno == "" {
		s.Catno = nil
	}
}
