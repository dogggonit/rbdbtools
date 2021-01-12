package database

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"rbdbtools/pkg/track"
	"rbdbtools/tools"
	"sort"
	"strings"
)

type Database struct {
	index     []track.Track
	files     map[string][]byte
	modified  bool
	bigEndian bool
}

func New(bigEndian bool) Database {
	return Database{
		index:     make([]track.Track, 0),
		bigEndian: bigEndian,
		modified:  true,
	}
}

func (d *Database) Add(tracks ...track.Track) {
	d.modified = true
	d.index = append(d.index, tracks...)
}

func (d *Database) Save(targetDir string) error {
	if targetDir == "" {
		return errors.New("target must be specified")
	} else if !tools.DirExists(targetDir) {
		return errors.New("target directory does not exist")
	}

	for k, v := range d.compile() {
		err := ioutil.WriteFile(path.Join(targetDir, k), v, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) Size() int {
	size := 0
	for _, v := range d.compile() {
		size += len(v)
	}
	return size
}

type Stats struct {
	Size   int
	Header HeaderData
}

type HeaderData struct {
	Size    int
	Entries int
}

func (d *Database) Entries() map[string]Stats {
	counts := make(map[string]Stats)

	for k, v := range d.compile() {
		s := Stats{
			Size: len(v),
			Header: HeaderData{
				Size:    int(tools.BytesNum(v[4:8], d.bigEndian)),
				Entries: int(tools.BytesNum(v[8:12], d.bigEndian)),
			},
		}

		switch k {
		case "database_0.tcd":
			counts["artists"] = s
		case "database_1.tcd":
			counts["albums"] = s
		case "database_2.tcd":
			counts["genres"] = s
		case "database_3.tcd":
			counts["tracks"] = s
		case "database_5.tcd":
			counts["composers"] = s
		case "database_6.tcd":
			counts["comments"] = s
		case "database_7.tcd":
			counts["album artists"] = s
		case "database_8.tcd":
			counts["groupings"] = s
		case "database_idx.tcd":
			counts["index"] = s
		}
	}

	return counts
}

func (d *Database) Tracks() []track.Track {
	return d.index
}

func (d *Database) sort() {
	ut := "<Untagged>"
	sort.Slice(d.index, func(i, j int) bool {
		s1 := strings.ToLower(d.index[i].Artist)
		if s1 == ut {
			return true
		}
		s2 := strings.ToLower(d.index[j].Artist)

		if s1 == s2 {
			i1 := d.index[i].Year
			i2 := d.index[j].Year

			if i1 == i2 {
				s1 = strings.ToLower(d.index[i].Album)
				if s1 == ut {
					return true
				}
				s2 = strings.ToLower(d.index[j].Album)

				if s1 == s2 {
					i1 = d.index[i].Track
					i2 = d.index[j].Track

					if i1 == i2 {
						s1 = strings.ToLower(d.index[i].Title)
						if s1 == ut {
							return true
						}
						s2 = strings.ToLower(d.index[j].Title)

						return s1 < s2 // By track name
					}
					return i1 < i2 // By track number
				}
				return s1 < s2 // By album
			}
			return i1 < i2 // By year
		}
		return s1 < s2 // By artist
	})
}
