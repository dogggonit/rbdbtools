package decoder

type DecodedDatabases struct {
	Tags      map[string]TagCache
	Index     IndexHeader
	BigEndian bool
}

type Header struct {
	Database string `csv:"database_name"`
	Filename string `csv:"filename"`
	Version  string `csv:"database_version"`
	Size     int32  `csv:"file_size"`
	Entries  int32  `csv:"number_entries"`
}

type TagCache struct {
	Header  Header
	Entries []TagCacheEntry
}

type TagCacheEntry struct {
	Offset   string `csv:"offset"`
	Size     int32  `csv:"data_length"`
	Index    int32  `csv:"index_value"`
	Data     string `csv:"data"`
	PaddedXs int    `csv:"padding"`
}

type IndexHeader struct {
	Header         Header
	Serial         int32 `csv:"serial"`
	CommitId       int32 `csv:"commit_id"`
	Dirty          bool  `csv:"dirty"`
	EntriesOffsets []IndexEntry
	EntriesTags    []IndexEntry
}

type IndexEntry struct {
	Index           int    `csv:"index"`
	Artist          string `csv:"artist"`
	Album           string `csv:"album"`
	Genre           string `csv:"genre"`
	Title           string `csv:"title"`
	Filename        string `csv:"filename"`
	Composer        string `csv:"composer"`
	Comment         string `csv:"comment"`
	AlbumArtist     string `csv:"album_artist"`
	Grouping        string `csv:"grouping"`
	Year            int32  `csv:"year"`
	DiscNumber      int32  `csv:"disc_number"`
	TrackNumber     int32  `csv:"track_number"`
	Bitrate         int32  `csv:"bitrate"`
	Length          int32  `csv:"length"`
	PlayCount       int32  `csv:"play_count"`
	Rating          int32  `csv:"rating"`
	PlayTime        int32  `csv:"playtime"`
	LastPlayed      int32  `csv:"last_played"`
	CommitId        int32  `csv:"commit_id"`
	Mtime           int32  `csv:"mtime"`
	LastElapsed     int32  `csv:"last_elapsed"`
	LastOffset      int32  `csv:"last_offset"`
	Flags           string `csv:"flags"`
	FlagDeleted     bool   `csv:"FLAG_DELETED"`
	FlagDirty       bool   `csv:"FLAG_DIRTY"`
	FlagTrackNumGen bool   `csv:"FLAG_TRKNUMGEN"`
	FlagResurrected bool   `csv:"FLAG_RESURRECTED"`
}

func (db *DecodedDatabases) GetHeaders() []Header {
	headers := []Header{db.Index.Header}

	for _, v := range db.Tags {
		headers = append(headers, v.Header)
	}

	return headers
}

func (db *DecodedDatabases) GetIndexHeader() []Header {
	return []Header{db.Index.Header}
}

func (db *DecodedDatabases) GetIndexTags() []IndexEntry {
	return db.Index.EntriesTags
}

func (db *DecodedDatabases) GetIndexOffsets() []IndexEntry {
	return db.Index.EntriesOffsets
}

func (db *DecodedDatabases) GetEntries(database string) []TagCacheEntry {
	return db.Tags[database].Entries
}

func (db *DecodedDatabases) GetDatabases() []string {
	dbs := make([]string, 0)
	for k := range db.Tags {
		dbs = append(dbs, k)
	}
	return dbs
}
