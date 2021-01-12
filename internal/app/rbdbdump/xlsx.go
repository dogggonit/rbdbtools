package rbdbdump

import (
	"errors"
	"github.com/tealeg/xlsx"
	"path"
	"rbdbtools/pkg/decoder"
	"reflect"
)

const (
	xlsxOut = "tagcache.xlsx"
)

func toXlsx(dbPath string, outPath string) {
	databases, err := decoder.DecodeDatabases(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	spreadsheet := xlsx.NewFile()

	headers := databases.GetHeaders()
	headersSheet, err := spreadsheet.AddSheet(headersCSV)
	if err != nil {
		log.Fatal(err)
	}
	err = addToSheet(headersSheet, headers)
	if err != nil {
		log.Fatal(err)
	}

	indexHeader := databases.GetIndexHeader()
	indexHeaderSheet, err := spreadsheet.AddSheet(indexHeaderCSV)
	if err != nil {
		log.Fatal(err)
	}
	err = addToSheet(indexHeaderSheet, indexHeader)
	if err != nil {
		log.Fatal(err)
	}

	indexEntries := databases.GetIndexOffsets()
	indexOffsetsSheet, err := spreadsheet.AddSheet(indexCSV)
	if err != nil {
		log.Fatal(err)
	}
	err = addToSheet(indexOffsetsSheet, indexEntries)
	if err != nil {
		log.Fatal(err)
	}

	indexEntries = databases.GetIndexTags()
	indexTagsSheet, err := spreadsheet.AddSheet(indexTagsCSV)
	if err != nil {
		log.Fatal(err)
	}
	err = addToSheet(indexTagsSheet, indexEntries)
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range databases.GetDatabases() {
		tagsSheet, err := spreadsheet.AddSheet(s)
		if err != nil {
			log.Fatal(err)
		}
		err = addToSheet(tagsSheet, databases.GetEntries(s))
		if err != nil {
			log.Fatal(err)
		}
	}

	err = spreadsheet.Save(path.Join(outPath, xlsxOut))
	if err != nil {
		log.Fatal(err)
	}
}

func addToSheet(sheet *xlsx.Sheet, elems interface{}) error {
	info := reflect.ValueOf(elems)
	if reflect.TypeOf(elems).Kind() != reflect.Slice {
		return errors.New("not a slice")
	}

	row := sheet.AddRow()
	for i, headers := 0, reflect.TypeOf(elems).Elem(); i < headers.NumField(); i++ {
		if tag, ok := headers.Field(i).Tag.Lookup("csv"); ok {
			row.AddCell().SetString(tag)
		}
	}

	for i := 0; i < info.Len(); i++ {
		row = sheet.AddRow()
		for j, e := 0, info.Index(i); j < e.NumField(); j++ {
			cell := row.AddCell()
			if _, ok := e.Type().Field(j).Tag.Lookup("csv"); ok {
				switch e.Type().Field(j).Type.Kind() {
				case reflect.Int32:
					fallthrough
				case reflect.Int:
					cell.SetInt64(e.Field(j).Int())
				case reflect.String:
					cell.SetString(e.Field(j).String())
				case reflect.Bool:
					cell.SetBool(e.Field(j).Bool())
				default:
					return errors.New("invalid type " + e.Type().Field(j).Type.Kind().String())
				}
			}
		}
	}

	return nil
}
