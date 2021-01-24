package xlsxdb

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"os"
	"path"
	"rbdbtools/pkg/database"
	"rbdbtools/tools"
	"strings"
)

var (
	HeadersColumns  = []string{"database_name", "filename", "database_version", "file_size_bytes", "number_entries", "serial", "commit_id", "dirty"}
	TagCacheColumns = []string{"offset", "data_length", "index", "data", "padding_Xs"}
	IndexColumns    = []string{"index", "artist", "album", "genre", "title", "filename", "composer", "comment", "album_artist", "grouping", "year", "disc_number", "track_number", "bitrate", "length", "play_count", "rating", "playtime", "last_played", "commit_id", "mtime_unix_ms", "last_elapsed", "last_offset", "flags", "FLAG_DELETED", "FLAG_DIRCACHE", "FLAG_DIRTYNUM", "FLAG_TRKNUMGEN", "FLAG_RESURRECTED", "flag_high"}
)

type XlsxDB struct {
	Xlsx *xlsx.File
	CSV  map[string][]byte
}

func New(db *database.Database) (*XlsxDB, error) {
	hexify := func(i int32) string {
		return fmt.Sprintf("0x%08X", i)
	}

	doc := &XlsxDB{
		Xlsx: xlsx.NewFile(),
		CSV:  make(map[string][]byte),
	}

	var tags [database.NumStringEntries]*xlsx.Sheet

	headers, err := doc.Xlsx.AddSheet("Headers")
	if err != nil {
		return nil, err
	}

	index, err := doc.Xlsx.AddSheet("Index")
	if err != nil {
		return nil, err
	}

	indexOffsets, err := doc.Xlsx.AddSheet("Index_Offsets")
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(tags); i++ {
		tags[i], err = doc.Xlsx.AddSheet(database.TagCache(i).String())
		if err != nil {
			return nil, err
		}
	}

	row := headers.AddRow()
	for _, col := range HeadersColumns {
		row.AddCell().SetString(col)
	}

	if err = database.ForEachTagCache(func(cache database.TagCache) error {
		row = headers.AddRow()

		header, err := db.GetHeader(cache)
		if err != nil {
			return err
		}

		row.AddCell().SetString(cache.String())
		row.AddCell().SetString(cache.Filename())
		row.AddCell().SetString(hexify(header.Magic))

		if cache == database.Index {
			idx, err := db.GetIdxHeader()
			if err != nil {
				return err
			}

			row.AddCell().SetString(hexify(idx.Serial))
			row.AddCell().SetString(hexify(idx.CommitId))
			row.AddCell().SetBool(idx.Dirty != 0)

			entries, err := db.GetIndexEntries()
			if err != nil {
				return err
			}

			for i, sheets := range []*xlsx.Sheet{index, indexOffsets} {
				row = sheets.AddRow()
				for _, col := range IndexColumns {
					row.AddCell().SetString(col)
				}

				for _, e := range entries {
					row = sheets.AddRow()

					row.AddCell().SetInt64(int64(e.Index))

					for j := 0; i < database.NumStringEntries; i++ {
						switch i {
						case 0:
							row.AddCell().SetString(e.Tags[j])
						case 1:
							row.AddCell().SetString(hexify(e.IntEntries[j]))
						}
					}

					for i := int(database.Year); i < database.NumIdxEntries; i++ {
						row.AddCell().SetInt64(int64(e.IntEntries[i]))
					}

					row.AddCell().SetInt64(int64(e.Int32Flag()))
					row.AddCell().SetBool(e.Flag.Deleted)
					row.AddCell().SetBool(e.Flag.DirCache)
					row.AddCell().SetBool(e.Flag.DirtyNum)
					row.AddCell().SetBool(e.Flag.TrkNumGen)
					row.AddCell().SetBool(e.Flag.Resurrected)
					row.AddCell().SetInt64(int64(e.Flag.High))
				}
			}

		} else {
			row = tags[cache].AddRow()
			for _, col := range TagCacheColumns {
				row.AddCell().SetString(col)
			}

			tagCache, err := db.GetAllTags(cache)
			if err != nil {
				return err
			}

			for _, tag := range tagCache {
				row = tags[cache].AddRow()

				row.AddCell().SetString(hexify(int32(tag.Offset)))
				row.AddCell().SetInt64(int64(tag.Length))
				row.AddCell().SetInt64(int64(tag.IdxId))
				row.AddCell().SetString(tag.TagData)
				row.AddCell().SetInt64(int64(int(tag.Length) - 1 - len(tag.TagData)))
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	for _, v := range doc.Xlsx.Sheets {
		cw := newCsvWriter()
		for _, r := range v.Rows {
			var vals []string
			for _, c := range r.Cells {
				str, err := c.FormattedValue()
				if err != nil {
					return nil, err
				}

				vals = append(vals, str)
			}

			if err = cw.cw.Write(vals); err != nil {
				return nil, err
			}
		}

		cw.cw.Flush()
		if err = cw.cw.Error(); err != nil {
			return nil, err
		}

		doc.CSV[v.Name] = cw.b
	}

	return doc, nil
}

func (xdb *XlsxDB) WriteXlsx(filename string) error {
	if ext := strings.Split(filename, "."); len(ext) < 2 && ext[1] != "xlsx" {
		return errors.New("filename is not a '.xlsx' file")
	}

	parts := strings.Split(filename, string(os.PathSeparator))
	dirname := path.Join(parts[:len(parts)-1]...)

	if !tools.DirExists(dirname) {
		err := os.MkdirAll(dirname, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return xdb.Xlsx.Save(filename)
}

func (xdb *XlsxDB) WriteCSV(dirname string) error {
	if tools.FileExists(dirname) {
		return errors.New("input must be a directory")
	}

	if !tools.DirExists(dirname) {
		err := os.MkdirAll(dirname, os.ModePerm)
		if err != nil {
			return err
		}
	}

	for k, v := range xdb.CSV {
		err := ioutil.WriteFile(path.Join(dirname, k+".csv"), v, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

type csvWriter struct {
	b  []byte
	cw *csv.Writer
}

func newCsvWriter() (cw csvWriter) {
	cw.cw = csv.NewWriter(&cw)
	return
}

func (cw *csvWriter) Write(p []byte) (n int, err error) {
	cpy := make([]byte, 0, len(p))
	copy(cpy, p)
	cw.b = append(cw.b, cpy...)
	return len(cpy), nil
}
