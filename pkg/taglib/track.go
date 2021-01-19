package taglib

func (t *TagLib) Artist() string {
	return t.artist
}

func (t *TagLib) Album() string {
	return t.album
}

func (t *TagLib) Genre() string {
	return t.genre
}

func (t *TagLib) Title() string {
	return t.title
}

func (t *TagLib) Filename() string {
	return t.filename
}

func (t *TagLib) SetFilename(filename string) {
	t.filename = filename
}

func (t *TagLib) Composer() string {
	return t.composer
}

func (t *TagLib) Comment() string {
	return t.comment
}

func (t *TagLib) AlbumArtist() string {
	return t.albumArtist
}

func (t *TagLib) Grouping() string {
	return t.grouping
}

func (t *TagLib) Year() int32 {
	return t.year
}

func (t *TagLib) Disc() int32 {
	return t.disc
}

func (t *TagLib) Track() int32 {
	return t.track
}

func (t *TagLib) Bitrate() int32 {
	return t.bitrate
}

func (t *TagLib) Length() int32 {
	return t.length
}

func (t *TagLib) Mtime() int32 {
	return t.mtime
}
