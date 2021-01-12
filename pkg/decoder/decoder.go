package decoder

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"rbdbtools/tools"
)

type offsetCache map[string]map[int32]string

func DecodeDatabases(dbPath string) (*DecodedDatabases, error) {
	databases := make(map[string][]byte)

	for k := range databaseToName {
		if db, err := ioutil.ReadFile(path.Join(dbPath, k)); err != nil {
			return nil, err
		} else if len(db) < 12 {
			return nil, InvalidHeaderError
		} else {
			databases[k] = db
		}
	}

	bigEndian, err := isBigEndian(databases)
	if err != nil {
		return nil, err
	}

	decoded := DecodedDatabases{
		Tags:      make(map[string]TagCache),
		BigEndian: bigEndian,
	}

	for k, v := range databases {
		if name := databaseToName[k]; name == index {
			if len(v) < 24 {
				return nil, InvalidHeaderError
			}
			decoded.Index = IndexHeader{
				Header:   decodeHeader(v, name, bigEndian),
				Serial:   tools.BytesNum(v[12:16], bigEndian),
				CommitId: tools.BytesNum(v[16:20], bigEndian),
				Dirty:    tools.BytesNum(v[20:24], bigEndian) != 0,
			}
		} else {
			decoded.Tags[name] = TagCache{
				Header: decodeHeader(v, name, bigEndian),
			}
		}
	}

	offsets, tags := decodeTagCaches(bigEndian, databases)
	for k, v := range decoded.Tags {
		decoded.Tags[k] = TagCache{
			Header:  v.Header,
			Entries: tags[k],
		}
	}
	decoded.Index.EntriesTags, decoded.Index.EntriesOffsets = decodeIndexEntries(databases[indexDB], bigEndian, offsets)

	return &decoded, nil
}

func isBigEndian(databases map[string][]byte) (bool, error) {
	var bigEndian *bool
	for _, v := range databases {
		thisDBIsBigEndian := string(v[0]) == "T"
		if bigEndian != nil && *bigEndian != thisDBIsBigEndian {
			return false, errors.New("database endian doesn't match between files")
		} else {
			bigEndian = new(bool)
		}
		*bigEndian = thisDBIsBigEndian
	}
	if bigEndian == nil {
		return false, errors.New("databases weren't loaded correctly")
	}
	return *bigEndian, nil
}

func decodeHeader(database []byte, dbName string, bigEndian bool) Header {
	ver := database[:4]
	size := database[4:8]
	entries := database[8:12]

	return Header{
		Database: dbName,
		Filename: nameToDatabase[dbName],
		Version:  fmt.Sprintf("0x%08X", tools.BytesNum(ver, bigEndian)),
		Size:     tools.BytesNum(size, bigEndian),
		Entries:  tools.BytesNum(entries, bigEndian),
	}
}

func decodeTagCaches(bigEndian bool, bytes map[string][]byte) (offsetCache, map[string][]TagCacheEntry) {
	cache := make(offsetCache)
	entries := make(map[string][]TagCacheEntry)

	for v := range databaseToName {
		if v == indexDB {
			continue
		} else {
			entries[databaseToName[v]] = make([]TagCacheEntry, 0)
		}

		for i := 12; i < len(bytes[v]); {
			entry := TagCacheEntry{
				Offset: fmt.Sprintf("0x%08X", int32(i)),
				Size:   tools.BytesNum(bytes[v][i:i+4], bigEndian),
				Index:  tools.BytesNum(bytes[v][i+4:i+8], bigEndian),
			}

			for end := i + 8; end < i+8+int(entry.Size); end++ {
				if bytes[v][end] == 0x00 {
					entry.Data = string(bytes[v][i+8 : end])
					entry.PaddedXs = i + 8 + int(entry.Size) - (end + 1)
					break
				}
			}

			entries[databaseToName[v]] = append(entries[databaseToName[v]], entry)
			cache.put(databaseToName[v], int32(i), entry.Data)

			i += int(entry.Size) + 8
		}
	}

	return cache, entries
}

func decodeIndexEntries(index []byte, bigEndian bool, offsets offsetCache) ([]IndexEntry, []IndexEntry) {
	entriesTags := make([]IndexEntry, 0)
	entriesOffsets := make([]IndexEntry, 0)

	for i, idx := 24, 0; i < len(index); i, idx = i+(23*4), idx+1 {
		artistOffset := tools.BytesNum(index[i:i+4], bigEndian)
		albumOffset := tools.BytesNum(index[i+4:i+8], bigEndian)
		genreOffset := tools.BytesNum(index[i+8:i+12], bigEndian)
		titleOffset := tools.BytesNum(index[i+12:i+16], bigEndian)
		filenameOffset := tools.BytesNum(index[i+16:i+20], bigEndian)
		composerOffset := tools.BytesNum(index[i+20:i+24], bigEndian)
		commentOffset := tools.BytesNum(index[i+24:i+28], bigEndian)
		albumArtistOffset := tools.BytesNum(index[i+28:i+32], bigEndian)
		groupingOffset := tools.BytesNum(index[i+32:i+36], bigEndian)
		flag := tools.BytesNum(index[i+88:i+92], bigEndian)

		entryTag := IndexEntry{
			Index:           idx,
			Artist:          offsets.get(artists, artistOffset),
			Album:           offsets.get(albums, albumOffset),
			Genre:           offsets.get(genres, genreOffset),
			Title:           offsets.get(titles, titleOffset),
			Filename:        offsets.get(filenames, filenameOffset),
			Composer:        offsets.get(composers, composerOffset),
			Comment:         offsets.get(comments, commentOffset),
			AlbumArtist:     offsets.get(albumArtists, albumArtistOffset),
			Grouping:        offsets.get(groupings, groupingOffset),
			Year:            tools.BytesNum(index[i+36:i+40], bigEndian),
			DiscNumber:      tools.BytesNum(index[i+40:i+44], bigEndian),
			TrackNumber:     tools.BytesNum(index[i+44:i+48], bigEndian),
			Bitrate:         tools.BytesNum(index[i+48:i+52], bigEndian),
			Length:          tools.BytesNum(index[i+52:i+56], bigEndian),
			PlayCount:       tools.BytesNum(index[i+56:i+60], bigEndian),
			Rating:          tools.BytesNum(index[i+60:i+64], bigEndian),
			PlayTime:        tools.BytesNum(index[i+64:i+68], bigEndian),
			LastPlayed:      tools.BytesNum(index[i+68:i+72], bigEndian),
			CommitId:        tools.BytesNum(index[i+72:i+76], bigEndian),
			Mtime:           tools.BytesNum(index[i+76:i+80], bigEndian),
			LastElapsed:     tools.BytesNum(index[i+80:i+84], bigEndian),
			LastOffset:      tools.BytesNum(index[i+84:i+88], bigEndian),
			Flags:           fmt.Sprintf("0x%08X", flag),
			FlagDeleted:     flag&int32(1) != 0,
			FlagDirty:       flag&int32(1<<2) != 0,
			FlagTrackNumGen: flag&int32(1<<3) != 0,
			FlagResurrected: flag&int32(1<<4) != 0,
		}

		entryOffset := entryTag
		entryOffset.Artist = fmt.Sprintf("0x%08X", artistOffset)
		entryOffset.Album = fmt.Sprintf("0x%08X", albumOffset)
		entryOffset.Genre = fmt.Sprintf("0x%08X", genreOffset)
		entryOffset.Title = fmt.Sprintf("0x%08X", titleOffset)
		entryOffset.Filename = fmt.Sprintf("0x%08X", filenameOffset)
		entryOffset.Composer = fmt.Sprintf("0x%08X", composerOffset)
		entryOffset.Comment = fmt.Sprintf("0x%08X", commentOffset)
		entryOffset.AlbumArtist = fmt.Sprintf("0x%08X", albumArtistOffset)
		entryOffset.Grouping = fmt.Sprintf("0x%08X", groupingOffset)
		entryOffset.Flags = fmt.Sprintf("0x%08X", flag)

		entriesTags = append(entriesTags, entryTag)
		entriesOffsets = append(entriesOffsets, entryOffset)
	}

	return entriesTags, entriesOffsets
}

func (oc offsetCache) get(database string, offset int32) string {
	tcem, e := oc[database]
	if !e {
		return fmt.Sprintf("DATABASE %s DOES NOT EXIST", database)
	}
	tce, e := tcem[offset]
	if !e {
		return fmt.Sprintf("OFFSET 0x%08X DOES NOT HAVE A VALUE IN DATABASE %s", offset, database)
	}
	return tce
}

func (oc offsetCache) put(database string, offset int32, value string) {
	if _, e := oc[database]; !e {
		oc[database] = make(map[int32]string)
	}
	oc[database][offset] = value
}
