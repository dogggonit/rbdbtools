package database

import (
	"rbdbtools/tools"
	"sort"
	"strings"
)

const (
	artist = iota
	album
	genre
	title
	filename
	composer
	comment
	albumartist
	grouping
	year
	discnumber
	tracknumber
	bitrate
	length
	playcount
	rating
	playtime
	lastplayed
	commitid
	mtime
	lastelapsed
	lastoffset
	tagCount

	dbVer = 0x5443480F
)

type tagEntry struct {
	offset  int
	tag     string
	length  [4]byte
	id      [4]byte
	data    []byte
	padding []byte
}

type indexEntry struct {
	index       int32
	tags        [tagCount][4]byte
	flag        [4]byte
	tagPointers [grouping + 1]*tagEntry
}

type header struct {
	version  [4]byte
	size     [4]byte
	entries  [4]byte
	elements []tagEntry
}

type masterHeader struct {
	header   header
	serial   [4]byte
	commitId [4]byte
	dirty    [4]byte
	elements []indexEntry
}

func (d *Database) compile() map[string][]byte {
	// No need to compile, return
	if !d.modified {
		return d.files
	}

	// Temp database
	tags := make(map[int]map[string]*tagEntry)
	titles := make([]*tagEntry, 0)
	filenames := make([]*tagEntry, 0)

	// Initialize sets for each type of tag
	for i := 0; i <= grouping; i++ {
		tags[i] = make(map[string]*tagEntry)
	}

	// Initialize headers
	index := masterHeader{
		elements: make([]indexEntry, 0),
	}
	copy(index.header.version[:], tools.NumBytes(dbVer, d.bigEndian))
	copy(index.commitId[:], tools.NumBytes(1, d.bigEndian))

	headers := make([]header, grouping+1)
	for i := range headers {
		copy(headers[i].version[:], tools.NumBytes(dbVer, d.bigEndian))
	}

	// Sort tracks before inserting into index
	d.sort()

	// Add tracks to index
	for i, e := range d.index {
		// Initialize index entry
		ie := indexEntry{
			index: int32(i),
		}

		// Get string fields from track
		fields := map[int]string{
			artist:      e.Artist,
			album:       e.Album,
			genre:       e.Genre,
			title:       e.Title,
			filename:    e.Filename,
			composer:    e.Composer,
			comment:     e.Comment,
			albumartist: e.AlbumArtist,
			grouping:    e.Grouping,
		}

		for k, v := range fields {
			if entry, exists := tags[k][v]; exists {
				// Add pointer field pointer to index entry if already has been added
				// Title and Filename should never exists
				ie.tagPointers[k] = entry
			} else {
				// Get padding size
				l := 0
				if !(k == title || k == filename) {
					l = (len(v) + 1) % 8
					if l != 0 {
						l = 8 - l
					}
				}
				// Add tag entry to index
				ie.tagPointers[k] = &tagEntry{
					tag:     v,
					data:    []byte(v),
					padding: []byte("XXXXXXXXX")[:l],
				}
				copy(ie.tagPointers[k].length[:], tools.NumBytes(uint32(len(v)+l+1), d.bigEndian))

				// Add tag to appropriate database
				// Title and filename use their own lists, since their entries are not unique
				if k == title {
					titles = append(titles, ie.tagPointers[k])
				} else if k == filename {
					filenames = append(filenames, ie.tagPointers[k])
				} else {
					tags[k][v] = ie.tagPointers[k]
				}

				if k == title || k == filename {
					// For title and filename entries, add index to entry
					copy(ie.tagPointers[k].id[:], tools.NumBytes(uint32(i), d.bigEndian))
				} else {
					// Otherwise set index to 0xFFFFFFFF
					copy(ie.tagPointers[k].id[:], tools.MaxBytes())
				}

				copy(ie.tags[year][:], tools.NumBytes(e.Year, d.bigEndian))
				copy(ie.tags[discnumber][:], tools.NumBytes(e.Disc, d.bigEndian))
				copy(ie.tags[tracknumber][:], tools.NumBytes(e.Track, d.bigEndian))
				copy(ie.tags[bitrate][:], tools.NumBytes(e.Bitrate, d.bigEndian))
				copy(ie.tags[length][:], tools.NumBytes(e.Length, d.bigEndian))
				copy(ie.tags[mtime][:], tools.NumBytes(e.Mtime, d.bigEndian))
			}
		}

		// Add to index
		index.elements = append(index.elements, ie)
	}

	// Set index header infos
	copy(index.header.entries[:], tools.NumBytes(uint32(len(index.elements)), d.bigEndian))
	copy(index.header.size[:], tools.NumBytes(uint32((len(index.elements)*(tagCount+1)*4)+12), d.bigEndian))

	// Create individual databases
	for k1, v1 := range tags {
		var db []*tagEntry

		// Assign list to correct entry, else create a list and add unique entries
		if k1 == title {
			db = titles
		} else if k1 == filename {
			db = filenames
		} else {
			db = make([]*tagEntry, 0)
			for _, v2 := range v1 {
				db = append(db, v2)
			}
		}

		// Sort the databases
		sort.Slice(db, func(i, j int) bool {
			return strings.ToLower(db[i].tag) < strings.ToLower(db[j].tag)
		})

		// Assign offsets and add tags to databases
		offset := 12
		for _, e := range db {
			e.offset = offset
			headers[k1].elements = append(headers[k1].elements, *e)

			offset += 8 + len(e.data) + 1 + len(e.padding)
		}

		copy(headers[k1].entries[:], tools.NumBytes(uint32(len(db)), d.bigEndian))
		copy(headers[k1].size[:], tools.NumBytes(uint32(offset-12), d.bigEndian))
	}

	// Add offsets to index
	for i := range index.elements {
		for j := 0; j <= grouping; j++ {
			copy(index.elements[i].tags[j][:], tools.NumBytes(uint32(index.elements[i].tagPointers[j].offset), d.bigEndian))
		}
	}

	// Generate databases
	tagToFilename := map[int]string{
		artist:      "database_0.tcd",
		album:       "database_1.tcd",
		genre:       "database_2.tcd",
		title:       "database_3.tcd",
		filename:    "database_4.tcd",
		composer:    "database_5.tcd",
		comment:     "database_6.tcd",
		albumartist: "database_7.tcd",
		grouping:    "database_8.tcd",
	}
	d.files = make(map[string][]byte)

	for k, v := range tagToFilename {
		d.files[v] = headers[k].version[:]
		d.files[v] = append(d.files[v], headers[k].size[:]...)
		d.files[v] = append(d.files[v], headers[k].entries[:]...)
		for _, e := range headers[k].elements {
			d.files[v] = append(d.files[v], e.length[:]...)
			d.files[v] = append(d.files[v], e.id[:]...)
			d.files[v] = append(d.files[v], e.data...)
			d.files[v] = append(d.files[v], 0x00)
			d.files[v] = append(d.files[v], e.padding...)
		}
	}

	// Generate index
	idx := "database_idx.tcd"
	d.files[idx] = index.header.version[:]
	d.files[idx] = append(d.files[idx], index.header.size[:]...)
	d.files[idx] = append(d.files[idx], index.header.entries[:]...)
	d.files[idx] = append(d.files[idx], index.serial[:]...)
	d.files[idx] = append(d.files[idx], index.commitId[:]...)
	d.files[idx] = append(d.files[idx], index.dirty[:]...)
	for _, e := range index.elements {
		for _, t := range e.tags {
			d.files[idx] = append(d.files[idx], t[:]...)
		}
		d.files[idx] = append(d.files[idx], e.flag[:]...)
	}

	d.modified = false
	return d.files
}
