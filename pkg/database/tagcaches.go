package database

const (
	Artist = TagCache(iota)
	Album
	Genre
	Title
	Filename
	Composer
	Comment
	AlbumArtist
	Grouping
	Index
)

// TagCache represents one of the 10 databases rockbox uses.
type TagCache uint

// String returns the string representation of the given TagCache,
// or an empty string if it is not a valid TagCache.
func (st TagCache) String() string {
	switch st {
	case Artist:
		return "Artist"
	case Album:
		return "Album"
	case Genre:
		return "Genre"
	case Title:
		return "Title"
	case Filename:
		return "Filename"
	case Composer:
		return "Composer"
	case Comment:
		return "Comment"
	case AlbumArtist:
		return "AlbumArtist"
	case Grouping:
		return "Grouping"
	case Index:
		return "Index"
	default:
		return ""
	}
}

// String returns the filename representation of the given TagCache,
// or an empty string if it is not a valid TagCache.
func (st TagCache) Filename() string {
	switch st {
	case Artist:
		return "database_0.tcd"
	case Album:
		return "database_1.tcd"
	case Genre:
		return "database_2.tcd"
	case Title:
		return "database_3.tcd"
	case Filename:
		return "database_4.tcd"
	case Composer:
		return "database_5.tcd"
	case Comment:
		return "database_6.tcd"
	case AlbumArtist:
		return "database_7.tcd"
	case Grouping:
		return "database_8.tcd"
	case Index:
		return "database_idx.tcd"
	default:
		return ""
	}
}

// ForEachTagCache cycles through each database and allow a function to be run on it.
func ForEachTagCache(consumer func(cache TagCache) error) error {
	err := consumer(Index)
	if err != nil {
		return err
	}
	for i := 0; i < NumStringEntries; i++ {
		err := consumer(TagCache(i))
		if err != nil {
			return err
		}
	}
	return nil
}
