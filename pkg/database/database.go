package database

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
)

const (
	CurrentVersion = 0x5443480F
	Untagged       = "<Untagged>"
)

type Database struct {
	Databases [NumStringEntries][]byte
	Idx       []byte
}

type Header struct {
	Endian  Endian
	Magic   int32
	Size    int
	Entries int
}

type Idx struct {
	Header   Header
	Serial   int32
	CommitId int32
	Dirty    int32
}

type TagCacheFiles map[string][]byte

func New(endian Endian) Database {
	db := Database{}
	_ = ForEachTagCache(func(cache TagCache) error {
		if cache == Index {
			db.Idx, _ = (&Idx{
				Header: Header{
					Endian:  endian,
					Magic:   CurrentVersion,
					Size:    0,
					Entries: 0,
				},
				Serial:   0,
				CommitId: 1,
				Dirty:    0,
			}).MarshalBinary()
		} else {
			db.Databases[cache], _ = (&Header{
				Endian:  endian,
				Magic:   CurrentVersion,
				Size:    0,
				Entries: 0,
			}).MarshalBinary()
		}
		return nil
	})
	return db
}

func (tcf *TagCacheFiles) ToDatabase() (*Database, error) {
	db := &Database{}
	missing := make([]string, 0, NumStringEntries+1)
	noHeader := make([]string, 0, NumStringEntries+1)

	_ = ForEachTagCache(func(cache TagCache) error {
		d, e := map[string][]byte(*tcf)[cache.Filename()]
		if !e {
			db = nil
			missing = append(missing, cache.String())
		} else if cache == Index {
			if len(d) < 24 {
				noHeader = append(noHeader, cache.String())
			} else {
				db.Idx = make([]byte, len(d))
				copy(db.Idx, d)
			}
		} else {
			if len(d) < 12 {
				noHeader = append(noHeader, cache.String())
			} else {
				db.Databases[cache] = make([]byte, len(d))
				copy(db.Databases[cache], d)
			}
		}
		return nil
	})

	if len(missing) > 0 {
		return nil, errors.New(fmt.Sprintf("missing databases: %s", missing))
	}
	if len(noHeader) > 0 {
		return nil, errors.New(fmt.Sprintf("databases that have no header: %s", noHeader))
	}

	return db, nil
}

func (db *Database) ToTagCacheFiles() TagCacheFiles {
	f := make(map[string][]byte)
	f[Index.Filename()] = db.Idx
	f[Index.Filename()] = make([]byte, len(db.Idx))
	copy(f[Index.Filename()], db.Idx)
	for i := 0; i < NumStringEntries; i++ {
		f[TagCache(i).Filename()] = make([]byte, len(db.Databases[i]))
		copy(f[TagCache(i).Filename()], db.Databases[i])
	}
	return f
}

func GetTagCacheFiles(dir string) (TagCacheFiles, error) {
	tcf := make(map[string][]byte)

	if err := ForEachTagCache(func(cache TagCache) error {
		b, err := ioutil.ReadFile(path.Join(dir, cache.Filename()))
		if err != nil {
			return err
		}

		tcf[cache.Filename()] = b
		return nil
	}); err != nil {
		return nil, err
	}

	return tcf, nil
}

func (header *Header) UnmarshalBinary(data []byte) error {
	if len(data) < 12 {
		return errors.New("no header exists")
	}
	header.Endian = string(data[0]) == "T"
	header.Magic = header.Endian.BytesNum(data[:4])
	header.Size = int(header.Endian.BytesNum(data[4:8]))
	header.Entries = int(header.Endian.BytesNum(data[8:12]))
	return nil
}

func (header *Header) MarshalBinary() (data []byte, err error) {
	data = header.Endian.NumBytes(header.Magic)
	data = append(data, header.Endian.NumBytes(int32(header.Size))...)
	data = append(data, header.Endian.NumBytes(int32(header.Entries))...)
	return data, nil
}

func (idx *Idx) UnmarshalBinary(data []byte) error {
	if len(data) < 24 {
		return errors.New("no header exists")
	}
	_ = (&idx.Header).UnmarshalBinary(data)
	idx.Serial = idx.Header.Endian.BytesNum(data[12:16])
	idx.CommitId = idx.Header.Endian.BytesNum(data[16:20])
	idx.Dirty = idx.Header.Endian.BytesNum(data[20:24])
	return nil
}

func (idx *Idx) MarshalBinary() (data []byte, err error) {
	data, _ = idx.Header.MarshalBinary()
	data = append(data, idx.Header.Endian.NumBytes(idx.Serial)...)
	data = append(data, idx.Header.Endian.NumBytes(idx.CommitId)...)
	data = append(data, idx.Header.Endian.NumBytes(idx.Dirty)...)
	return data, nil
}

func (db *Database) GetHeader(t TagCache) (*Header, error) {
	if t > Index {
		return nil, errors.New("invalid database")
	}

	var header Header
	err := header.UnmarshalBinary(db.Databases[t])
	if err != nil {
		return nil, err
	}

	return &header, nil
}

func (db *Database) GetIdxHeader() (*Idx, error) {
	header := Idx{}
	err := header.UnmarshalBinary(db.Idx)
	if err != nil {
		return nil, err
	}
	return &header, nil
}
