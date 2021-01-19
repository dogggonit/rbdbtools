package database

import "time"

const (
	Year = IdxEntry(iota + Grouping)
	Disc
	Track
	Bitrate
	Length
	PlayCount
	Rating
	PlayTime
	LastPlayed
	CommitId
	Mtime
	LastElapsed
	LastOffset
	NumIdxEntries    = int(LastOffset + 1)
	NumStringEntries = int(Grouping + 1)
)

type IdxEntry uint

func (idx IdxEntry) String() string {
	switch idx {
	case Year:
		return "Year"
	case Disc:
		return "Disc"
	case Track:
		return "Track"
	case Bitrate:
		return "Bitrate"
	case Length:
		return "Length"
	case PlayCount:
		return "PlayCount"
	case Rating:
		return "Rating"
	case PlayTime:
		return "PlayTime"
	case LastPlayed:
		return "LastPlayed"
	case CommitId:
		return "CommitId"
	case Mtime:
		return "Mtime"
	case LastElapsed:
		return "LastElapsed"
	case LastOffset:
		return "LastOffset"
	default:
		if idx < Year {
			return TagCache(idx).String()
		}
		return ""
	}
}

type IndexEntry struct {
	Header     *Idx
	Index      int32
	Tags       [NumStringEntries]string
	IntEntries [NumIdxEntries]int32
	MTime      time.Time
	Flag       struct {
		Deleted     bool
		DirCache    bool
		DirtyNum    bool
		TrkNumGen   bool
		Resurrected bool
		High        uint16
	}
}

func (db *Database) GetIndexEntryByIndex(idx int) (*IndexEntry, error) {
	header := &Idx{}
	err := header.UnmarshalBinary(db.Idx)
	if err != nil {
		return nil, err
	}

	idx = 24 + (idx * 23 * 4)

	entry := &IndexEntry{
		Header: header,
		Index:  int32(idx),
	}

	for i, j := idx, 0; j < NumIdxEntries; i, j = i+4, j+1 {
		entry.IntEntries[j] = header.Header.Endian.BytesNum(db.Idx[i : i+4])
	}

	for i := 0; i < NumStringEntries; i++ {
		tag, err := db.GetTagFileEntryByOffset(TagCache(i), int(entry.IntEntries[i]))
		if err != nil {
			return nil, err
		}
		entry.Tags[i] = tag.TagData
	}

	flag := header.Header.Endian.BytesNum(db.Idx[idx+(22*4) : idx+(23*4)])
	entry.Flag.High = uint16(flag >> 16)
	entry.Flag.Deleted = flag&1 > 0
	entry.Flag.DirCache = flag&(1<<1) > 0
	entry.Flag.DirtyNum = flag&(1<<2) > 0
	entry.Flag.TrkNumGen = flag&(1<<3) > 0
	entry.Flag.Resurrected = flag&(1<<4) > 0

	entry.MTime = time.Unix(0, int64(entry.IntEntries[Mtime])*int64(1000000))
	return entry, nil
}

func (db *Database) GetIndexEntries() ([]IndexEntry, error) {
	header := &Idx{}
	err := header.UnmarshalBinary(db.Idx)
	if err != nil {
		return nil, err
	}

	entries := make([]IndexEntry, 0, header.Header.Entries)
	for i := 0; i < header.Header.Entries; i++ {
		entry, err := db.GetIndexEntryByIndex(i)
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}

	return entries, nil
}

func (db *Database) DefaultSortIndex() error {
	return db.SortIndex(func(e1, e2 IndexEntry) bool {
		// TODO
		return false
	})
}

func (db *Database) SortIndex(less func(e1, e2 IndexEntry) bool) error {
	entries, err := db.GetIndexEntries()
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return nil
	}

	// TODO

	return nil
}

func (i IndexEntry) Artist() string {
	return i.Tags[Artist]
}

func (i IndexEntry) Album() string {
	return i.Tags[Album]
}

func (i IndexEntry) Genre() string {
	return i.Tags[Genre]
}

func (i IndexEntry) Title() string {
	return i.Tags[Title]
}

func (i IndexEntry) Filename() string {
	return i.Tags[Filename]
}

func (i IndexEntry) SetFilename(_ string) {
	// No op
}

func (i IndexEntry) Composer() string {
	return i.Tags[Composer]
}

func (i IndexEntry) Comment() string {
	return i.Tags[Comment]
}

func (i IndexEntry) AlbumArtist() string {
	return i.Tags[AlbumArtist]
}

func (i IndexEntry) Grouping() string {
	return i.Tags[Grouping]
}

func (i IndexEntry) Year() int32 {
	return i.IntEntries[Year]
}

func (i IndexEntry) Disc() int32 {
	return i.IntEntries[Disc]
}

func (i IndexEntry) Track() int32 {
	return i.IntEntries[Track]
}

func (i IndexEntry) Bitrate() int32 {
	return i.IntEntries[Bitrate]
}

func (i IndexEntry) Length() int32 {
	return i.IntEntries[Length]
}

func (i IndexEntry) Mtime() int32 {
	return i.IntEntries[Mtime]
}
