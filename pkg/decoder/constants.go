package decoder

import "errors"

const (
	artists      = "artists"
	albums       = "albums"
	genres       = "genres"
	titles       = "titles"
	filenames    = "filenames"
	composers    = "composers"
	comments     = "comments"
	albumArtists = "albumArtists"
	groupings    = "groupings"
	index        = "index"

	artistsDB      = "database_0.tcd"
	albumsDB       = "database_1.tcd"
	genresDB       = "database_2.tcd"
	titlesDB       = "database_3.tcd"
	filenamesDB    = "database_4.tcd"
	composersDB    = "database_5.tcd"
	commentsDB     = "database_6.tcd"
	albumArtistsDB = "database_7.tcd"
	groupingsDB    = "database_8.tcd"
	indexDB        = "database_idx.tcd"
)

var (
	InvalidHeaderError = errors.New("header is not valid")
)

var (
	nameToDatabase = map[string]string{
		artists:      artistsDB,
		albums:       albumsDB,
		genres:       genresDB,
		titles:       titlesDB,
		filenames:    filenamesDB,
		composers:    composersDB,
		comments:     commentsDB,
		albumArtists: albumArtistsDB,
		groupings:    groupingsDB,
		index:        indexDB,
	}
	databaseToName = map[string]string{
		artistsDB:      artists,
		albumsDB:       albums,
		genresDB:       genres,
		titlesDB:       titles,
		filenamesDB:    filenames,
		composersDB:    composers,
		commentsDB:     comments,
		albumArtistsDB: albumArtists,
		groupingsDB:    groupings,
		indexDB:        index,
	}
)
