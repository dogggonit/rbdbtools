package database

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

type TagFileEntry struct {
	Header   *Header
	Offset   int
	Length   int32
	IdxId    int32
	TagData  string
	TagCache TagCache
}

func (db *Database) GetTagFileEntryByString(t TagCache, s string) (*TagFileEntry, error) {
	for i := 12; i < len(db.Databases[t]); {
		entry, err := db.GetTagFileEntryByOffset(t, i)
		if err != nil {
			return nil, err
		}

		if s == entry.TagData {
			return entry, nil
		}
		i += 8 + int(entry.Length)
	}

	return nil, errors.New(fmt.Sprintf("entry for %s not found", s))
}

func (db *Database) GetTagFileEntryByIndex(t TagCache, idx int) (*TagFileEntry, error) {
	for i, index := 0, 0; i < len(db.Databases[t]) && index < idx; index++ {
		entry, err := db.GetTagFileEntryByOffset(t, i)
		if err != nil {
			return nil, err
		}

		if index == idx {
			return entry, nil
		}
		i += 8 + int(entry.Length)
	}

	return nil, errors.New(fmt.Sprintf("could not find index %d", idx))
}

func (db *Database) GetTagFileEntryByOffset(t TagCache, offset int) (*TagFileEntry, error) {
	header, err := db.GetHeader(t)
	if err != nil {
		return nil, err
	}

	entry := TagFileEntry{
		Header:   header,
		Offset:   offset,
		TagCache: t,
	}

	err = entry.UnmarshalBinary(db.Databases[t][offset:])
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (db *Database) AddTag(t TagCache, tag string) (offset int) {
	header, err := db.GetHeader(t)
	if err != nil {
		return -1
	}
	offset = len(db.Databases[t])

	data := header.Endian.NumBytes(-1)

	data = append(data, []byte(tag)...)
	data = append(data, 0x00)
	if t != Filename {
		l := (len(tag) + 1) % 8
		if l != 0 {
			l = 8 - l
		}
		data = append(data, []byte("XXXXXXXXX")[:l]...)
	}
	data = append(header.Endian.NumBytes(int32(len(data)-4)), data...)
	return
}

func (db *Database) GetAllTags(t TagCache) ([]TagFileEntry, error) {
	header, err := db.GetHeader(t)
	if err != nil {
		return nil, err

	}

	entries := make([]TagFileEntry, 0, header.Entries)
	for i := 0; i < header.Entries; i++ {
		entry, err := db.GetTagFileEntryByIndex(t, i)
		if err != nil {
			return nil, err
		}

		if i > 0 {
			entry.Header = entries[i-1].Header
		}

		entries = append(entries, *entry)
	}

	return entries, nil
}

func (db *Database) DefaultSortTags(t TagCache) error {
	return db.SortTags(t, func(e1, e2 TagFileEntry) bool {
		if e1.TagData == Untagged {
			return true
		} else if e1.TagCache != e2.TagCache {
			return e1.TagData < e2.TagData
		} else if e1.TagCache == Filename {
			return e1.IdxId < e2.IdxId
		}
		return strings.ToLower(e1.TagData) < strings.ToLower(e2.TagData)
	})
}

func (db *Database) SortTags(t TagCache, less func(e1, e2 TagFileEntry) bool) error {
	if len(db.Idx) < 24 {
		return errors.New("no header in database file")
	}

	entries, err := db.GetAllTags(t)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return nil
	}

	header := entries[0].Header

	sort.Slice(entries, func(i, j int) bool {
		return less(entries[i], entries[j])
	})

	newDB := append(make([]byte, 0, len(db.Databases[t])), db.Databases[t][:12]...)

	idx := make([]byte, len(db.Idx)-24)
	copy(idx, db.Idx[:24])

	completed := false
	defer func() {
		if completed {
			db.Databases[t] = newDB
			db.Idx = append(db.Idx[:24], idx...)
		}
	}()

	mapOffsets := make(map[int]int)
	for _, e := range entries {
		mapOffsets[e.Offset] = len(newDB)
		b, err := e.MarshalBinary()
		if err != nil {
			return err
		}
		newDB = append(newDB, b...)
	}

	for i := 0; i < len(idx); i += 23 * 4 {
		oldOffset := int(header.Endian.BytesNum(idx[i+(int(t)*4) : i+(int(t)*4)+4]))
		if newOffset, exists := mapOffsets[oldOffset]; exists {
			n := header.Endian.NumBytes(int32(newOffset))
			idx = append(idx[:i+(int(t)*4)], append(n, idx[i+(int(t)*4)+4:]...)...)
		} else {
			return errors.New(fmt.Sprintf("index for %s does not have offset 0x%08X", t.String(), oldOffset))
		}
	}

	completed = true
	return nil
}

func (tfe *TagFileEntry) MarshalBinary() (data []byte, err error) {
	if tfe.Header == nil {
		return nil, errors.New("no header set in TagFileEntry")
	}

	if tfe.TagCache == Filename || tfe.TagCache == Title {
		data = tfe.Header.Endian.NumBytes(tfe.IdxId)
	} else {
		data = []byte{0xFF, 0xFF, 0xFF, 0xFF}
	}

	l := 0
	if tfe.TagCache != Filename {
		l = (len(tfe.TagData) + 1) % 8
		if l != 0 {
			l = 8 - l
		}
	}
	tag := append([]byte(tfe.TagData), append([]byte{0x00}, []byte("XXXXXXXXX")[:l]...)...)
	return append(tfe.Header.Endian.NumBytes(int32(len(tag))), append(data, tag...)...), nil
}

func (tfe *TagFileEntry) UnmarshalBinary(data []byte) error {
	if tfe.Header == nil {
		return errors.New("no header set in TagFileEntry")
	}

	if len(data) < 8 {
		return errors.New("invalid entry data")
	}

	tfe.Length = tfe.Header.Endian.BytesNum(data[:4])
	tfe.IdxId = tfe.Header.Endian.BytesNum(data[4:8])

	tag := data[8:int(math.Min(8+float64(tfe.Length), float64(len(data))))]

	for i, b := range tag {
		if b == 0x00 {
			tag = tag[:i]
			break
		}
	}

	if len(tag) == len(data)-8 && tfe.Length != 0 {
		return errors.New("string doesn't end in null byte")
	}

	tfe.TagData = string(tag)

	return nil
}
