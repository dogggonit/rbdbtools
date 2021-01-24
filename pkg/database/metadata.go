package database

// MetaData is used to import and export a track to and from the rockbox database.
type MetaData interface {
	Artist() string
	Album() string
	Genre() string
	Title() string
	Filename() string
	SetFilename(filename string)
	Composer() string
	Comment() string
	AlbumArtist() string // Default to Artist if the audio does not have this field
	Grouping() string    // Default to Title if the audio does not have this field
	Year() int32
	Disc() int32
	Track() int32
	Bitrate() int32
	Length() int32 // In milliseconds
	Mtime() int32  // Unix time in milliseconds
}
