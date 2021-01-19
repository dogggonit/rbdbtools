package database

type TrackData interface {
	Artist() string
	Album() string
	Genre() string
	Title() string
	Filename() string
	SetFilename(filename string)
	Composer() string
	Comment() string
	AlbumArtist() string
	Grouping() string
	Year() int32
	Disc() int32
	Track() int32
	Bitrate() int32
	Length() int32
	Mtime() int32
}
