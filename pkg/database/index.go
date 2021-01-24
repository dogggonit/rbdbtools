package database

import (
	"errors"
	"sort"
	"strings"
	"time"
)

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

	entry.MTime = millisecondsToTime(entry.IntEntries[Mtime])
	return entry, nil
}

func millisecondsToTime(ms int32) time.Time {
	return time.Unix(0, int64(ms)*int64(1000000))
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
		s1 := strings.ToLower(e1.AlbumArtist())
		s2 := strings.ToLower(e2.AlbumArtist())
		if s1 == s2 {
			if e1.Year() == e2.Year() {
				s1 = strings.ToLower(e1.Album())
				s2 = strings.ToLower(e2.Album())
				if s1 == s2 {
					if e1.Track() == e2.Track() {
						s1 = strings.ToLower(e1.Grouping())
						s2 = strings.ToLower(e2.Grouping())
						return s1 < s2
					}
					return e1.Track() < e2.Track()
				}
				return s1 < s2
			}
			return e1.Year() < e2.Year()
		}
		return s1 < s2
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

	dbBak := Database{}
	dbBak.Idx = make([]byte, 24, len(db.Idx))
	copy(dbBak.Idx, db.Idx[:24])
	dbBak.Databases[Filename] = make([]byte, len(db.Databases[Filename]))
	copy(dbBak.Databases[Filename], db.Databases[Filename])
	dbBak.Databases[Title] = make([]byte, len(db.Databases[Title]))
	copy(dbBak.Databases[Title], db.Databases[Title])

	sort.Slice(entries, func(i, j int) bool {
		return less(entries[i], entries[j])
	})

	for i, e := range entries {
		b, err := e.MarshalBinary()
		if err != nil {
			return err
		}
		dbBak.Idx = append(dbBak.Idx, b...)

		idx := e.Header.Header.Endian.NumBytes(int32(i))

		o := e.IntEntries[Title]
		dbBak.Databases[Title] = append(dbBak.Databases[Title][:o+4], append(idx, dbBak.Databases[Title][o+8:]...)...)

		o = e.IntEntries[Filename]
		dbBak.Databases[Filename] = append(dbBak.Databases[Filename][:o+4], append(idx, dbBak.Databases[Filename][o+8:]...)...)
	}

	err = dbBak.DefaultSortTags(Filename)
	if err != nil {
		return err
	}

	db.Idx = dbBak.Idx
	db.Databases[Title] = dbBak.Databases[Title]
	db.Databases[Filename] = dbBak.Databases[Filename]

	return nil
}

func (db *Database) AppendTracks(tracks ...MetaData) ([]IndexEntry, error) {
	idx, err := db.GetIdxHeader()
	if err != nil {
		return nil, err
	}

	var entries []IndexEntry
	for _, t := range tracks {
		idxEntry, err := db.AppendTrack(t)
		if err != nil {
			return nil, err
		}

		idxEntry.Header = idx
		entries = append(entries, *idxEntry)
	}

	newIdx, err := db.GetIdxHeader()
	if err != nil {
		return nil, err
	}

	idx.Header = newIdx.Header

	idx.CommitId++
	ci := idx.Header.Endian.NumBytes(idx.CommitId)
	db.Idx = append(db.Idx[:16], append(ci, db.Idx[20:]...)...)

	return entries, nil
}

// AppendTrack adds a track to the database.
func (db *Database) AppendTrack(t MetaData) (*IndexEntry, error) {
	header, err := db.GetIdxHeader()
	if err != nil {
		return nil, err
	}

	idxEntry := IndexEntry{
		Header: header,
	}

	stringTags := []string{t.Artist(), t.Album(), t.Genre(), t.Title(), t.Filename(), t.Composer(), t.Comment(), t.AlbumArtist(), t.Grouping()}

	for i := 0; i < NumIdxEntries; i++ {
		switch {
		case TagCache(i) <= Grouping:
			tag, err := db.AppendTag(TagCache(i), stringTags[i])

			if err != nil {
				return nil, err
			}

			idxEntry.Tags[i] = tag.TagData
			idxEntry.IntEntries[i] = int32(tag.Offset)
		case IdxEntry(i) == Year:
			idxEntry.IntEntries[i] = t.Year()
		case IdxEntry(i) == Disc:
			idxEntry.IntEntries[i] = t.Disc()
		case IdxEntry(i) == Track:
			idxEntry.IntEntries[i] = t.Track()
		case IdxEntry(i) == Bitrate:
			idxEntry.IntEntries[i] = t.Bitrate()
		case IdxEntry(i) == Length:
			idxEntry.IntEntries[i] = t.Length()
		case IdxEntry(i) == Mtime:
			idxEntry.IntEntries[i] = t.Mtime()
		}
	}

	idxEntry.MTime = millisecondsToTime(idxEntry.IntEntries[Mtime])
	idxEntry.Index = int32(header.Header.Entries) + 1

	data, err := idxEntry.MarshalBinary()
	if err != nil {
		return nil, err
	}

	db.Idx = append(db.Idx, data...)

	return &idxEntry, nil
}

func (db *Database) DeleteIndexEntry(idx int) (*IndexEntry, error) {
	e, err := db.GetIndexEntryByIndex(idx)
	if err != nil {
		return nil, err
	}

	dbBak := Database{}
	dbBak.Idx = make([]byte, len(db.Idx))
	copy(dbBak.Idx, db.Idx)
	dbBak.Databases[Filename] = make([]byte, len(db.Databases[Filename]))
	copy(dbBak.Databases[Filename], db.Databases[Filename])
	dbBak.Databases[Title] = make([]byte, len(db.Databases[Title]))
	copy(dbBak.Databases[Title], db.Databases[Title])

	newHeader := Idx{}
	err = newHeader.UnmarshalBinary(db.Idx)
	newHeader.Header.Entries--
	newHeader.Header.Size -= 23*4
	b, err := newHeader.MarshalBinary()
	if err != nil {
		return nil, err
	}
	dbBak.Idx = append(b, dbBak.Idx[24:]...)

	// TODO remove the index entry
	// TODO delete Title and Filename entries
	// TODO update index values in Title and Filename databases
	// TODO check other tags used, and if they are unused remove them
	// TODO reassign databases back to db

	return e, nil
}

func (ie *IndexEntry) MarshalBinary() (data []byte, err error) {
	if ie.Header == nil {
		return nil, errors.New("no header set in IndexEntry")
	}

	for _, e := range ie.IntEntries {
		data = append(data, ie.Header.Header.Endian.NumBytes(e)...)
	}

	data = append(data, ie.Header.Header.Endian.NumBytes(ie.Int32Flag())...)
	return
}

func (ie *IndexEntry) Int32Flag() (flag int32) {
	flag = int32(ie.Flag.High) << 16

	switch {
	case ie.Flag.Deleted:
		flag |= 0x1
	case ie.Flag.DirCache:
		flag |= 0x2
	case ie.Flag.DirtyNum:
		flag |= 0x4
	case ie.Flag.TrkNumGen:
		flag |= 0x8
	case ie.Flag.Resurrected:
		flag |= 0x10
	}

	return
}

func (ie *IndexEntry) Artist() string {
	return ie.Tags[Artist]
}

func (ie *IndexEntry) Album() string {
	return ie.Tags[Album]
}

func (ie *IndexEntry) Genre() string {
	return ie.Tags[Genre]
}

func (ie *IndexEntry) Title() string {
	return ie.Tags[Title]
}

func (ie *IndexEntry) Filename() string {
	return ie.Tags[Filename]
}

func (ie *IndexEntry) SetFilename(_ string) {
	// No op
}

func (ie *IndexEntry) Composer() string {
	return ie.Tags[Composer]
}

func (ie *IndexEntry) Comment() string {
	return ie.Tags[Comment]
}

func (ie *IndexEntry) AlbumArtist() string {
	return ie.Tags[AlbumArtist]
}

func (ie *IndexEntry) Grouping() string {
	return ie.Tags[Grouping]
}

func (ie *IndexEntry) Year() int32 {
	return ie.IntEntries[Year]
}

func (ie *IndexEntry) Disc() int32 {
	return ie.IntEntries[Disc]
}

func (ie *IndexEntry) Track() int32 {
	return ie.IntEntries[Track]
}

func (ie *IndexEntry) Bitrate() int32 {
	return ie.IntEntries[Bitrate]
}

func (ie *IndexEntry) Length() int32 {
	return ie.IntEntries[Length]
}

func (ie *IndexEntry) Mtime() int32 {
	return ie.IntEntries[Mtime]
}
