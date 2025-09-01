package discogs

type Artist struct {
	ID          int64  `xml:"id"`
	Name        string `xml:"name"`
	RealName    string `xml:"realname"`
	Profile     string `xml:"profile"`
	DataQuality string `xml:"data_quality"`

	URLs           []string `xml:"urls>url"`
	Aliases        []Name   `xml:"aliases>name"`
	NameVariations []string `xml:"namevariations>name"`
	Members        []Name   `xml:"members>name"`
	Groups         []Name   `xml:"groups>name"`
}

type Name struct {
	ID   int64  `xml:"id,attr"`
	Name string `xml:",chardata"`
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
	records := make([][]any, 0)
	for _, alias := range a.Aliases {
		records = append(records, []any{a.ID, alias.ID})
	}
	return records
}

func (a *Artist) ToMembershipsRecords() [][]any {
	records := make([][]any, 0)
	for _, member := range a.Members {
		records = append(records, []any{a.ID, member.ID})
	}
	return records
}

type Label struct {
	ID          int64      `xml:"id"`
	Name        string     `xml:"name"`
	ContactInfo string     `xml:"contactinfo"`
	Profile     string     `xml:"profile"`
	DataQuality string     `xml:"data_quality"`
	URLs        []string   `xml:"urls>url"`
	ParentLabel SubLabel   `xml:"parentlabel>label"`
	SubLabels   []SubLabel `xml:"sublabels>label"`
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
	ID   int64  `xml:"id,attr"`
	Name string `xml:",chardata"`
}

func (l *Label) ToRecord() []any {
	var parentID *int64
	if l.ParentLabel.ID != 0 {
		parentID = &l.ParentLabel.ID
	}
	return []any{
		l.ID,
		parentID,
		l.DataQuality,
		l.Name,
		l.Profile,
		l.ContactInfo,
		l.URLs,
	}
}

type Master struct {
	ID          int64          `xml:"id,attr"`
	Artists     []MasterArtist `xml:"artists>artist"`
	MainRelease int64          `xml:"mainrelease"`
	DataQuality string         `xml:"data_quality"`
	Videos      []Video        `xml:"videos>video"`
	Title       string         `xml:"title"`
	Year        int32          `xml:"year"`
	Genres      []string       `xml:"genres>genre"`
	Styles      []string       `xml:"styles>style"`
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
		m.MainRelease,
		m.DataQuality,
		m.Title,
		m.Year,
		m.Genres,
		m.Styles,
		m.Videos,
	}
}

func (m *Master) ToArtistsRecords() [][]any {
	artists := make([][]any, 0)
	for _, artist := range m.Artists {
		artists = append(artists, []any{
			m.ID,
			artist.ID,
			artist.Name,
			artist.Anv,
			artist.Join,
		})
	}

	return artists
}

type MasterArtist struct {
	ID   int64  `xml:"id"`
	Name string `xml:"name"`
	Anv  string `xml:"anv"`
	Join string `xml:"join"`
}

type Video struct {
	Src         string `xml:"src,attr" json:"src"`
	Duration    int32  `xml:"duration,attr" json:"duration"`
	Embed       string `xml:"embed,attr" json:"embed"`
	Title       string `xml:"title" json:"title"`
	Description string `xml:"description" json:"description"`
}

type Release struct {
	ID           int64           `xml:"id,attr"`
	MasterID     MasterID        `xml:"master_id"`
	Status       string          `xml:"status,attr"`
	Artists      []MasterArtist  `xml:"artists"`
	ExtraArtists []ExtraArtist   `xml:"extraartists>artist"`
	Title        string          `xml:"title"`
	Labels       []ReleaseLabel  `xml:"labels>label"`
	Country      string          `xml:"country"`
	Released     string          `xml:"released"`
	Notes        string          `xml:"notes"`
	DataQuality  string          `xml:"data_quality"`
	Genres       []string        `xml:"genres>genre"`
	Styles       []string        `xml:"styles>style"`
	Identifiers  []Identifier    `xml:"identifiers>identifier"`
	Videos       []Video         `xml:"videos>video"`
	Formats      []ReleaseFormat `xml:"formats>format"`
	Tracklist    []Track         `xml:"tracklist>track"`
	Companies    []Company       `xml:"companies>company"`
	Series       []Serie         `xml:"series>serie"`
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
	artists := make([][]any, 0)
	for _, artist := range r.Artists {
		artists = append(artists, []any{
			r.ID,
			artist.ID,
			artist.Name,
			artist.Anv,
			artist.Join,
		})
	}

	return artists
}

func (r *Release) ToRecord() []any {
	var masterID *int64
	if r.MasterID.ID != 0 {
		masterID = &r.MasterID.ID
	}

	return []any{
		r.ID,
		masterID,
		r.MasterID.IsMainRelease,
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
	extraArtists := make([][]any, 0)
	for _, artist := range r.ExtraArtists {
		extraArtists = append(extraArtists, []any{
			r.ID,
			artist.ID,
			artist.Name,
			artist.Anv,
			artist.Role,
		})
	}

	return extraArtists
}

func (r *Release) ToLabelsRecords() [][]any {
	labels := make([][]any, 0)
	for _, label := range r.Labels {
		labels = append(labels, []any{
			r.ID,
			label.ID,
			label.Name,
			label.Catno,
		})
	}

	return labels
}

type MasterID struct {
	IsMainRelease bool  `xml:"is_main_release,attr"`
	ID            int64 `xml:",chardata"`
}

type Track struct {
	Artists  []Artist `xml:"artists>artist"`
	Position string   `xml:"position,attr"`
	Title    string   `xml:"title"`
	Duration string   `xml:"duration"`
}

type Identifier struct {
	Type        string `xml:"type,attr"`
	Description string `xml:"description,attr"`
	Value       string `xml:"value,attr"`
}

type ExtraArtist struct {
	ID   int64  `xml:"id"`
	Name string `xml:"name"`
	Anv  string `xml:"anv"`
	Role string `xml:"role"`
}

type ReleaseLabel struct {
	ID    int64  `xml:"id,attr"`
	Name  string `xml:"name,attr"`
	Catno string `xml:"catno,attr"`
}

type Company struct {
	ID             int64  `xml:"id"`
	Name           string `xml:"name"`
	EntityType     int64  `xml:"entity_type"`
	EntityTypeName string `xml:"entity_type_name"`
	ResourceURL    string `xml:"resource_url"`
}

type ReleaseFormat struct {
	Name         string   `xml:"name,attr"`
	Qty          string   `xml:"qty,attr"`
	Text         string   `xml:"text,attr"`
	Descriptions []string `xml:"descriptions>description"`
}

type Serie struct {
	Name  string `xml:"name,attr"`
	ID    int64  `xml:"id,attr"`
	Catno string `xml:"catno,attr"`
}
