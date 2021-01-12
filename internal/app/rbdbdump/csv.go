package rbdbdump

import (
	"fmt"
	"github.com/gocarina/gocsv"
	"os"
	"path"
	"rbdbtools/pkg/decoder"
	"rbdbtools/pkg/logger"
)

const (
	headersCSV     = "headers"
	indexHeaderCSV = "index_header"
	indexCSV       = "index"
	indexTagsCSV   = "indexTags"
)

var (
	log = logger.New()
)

func csv(dbPath string, outPath string) {
	databases, err := decoder.DecodeDatabases(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	headers := databases.GetHeaders()
	err = writeCSV(&headers, path.Join(outPath, headersCSV+".csv"))
	if err != nil {
		log.Error(err)
	}
	indexHeader := databases.GetIndexHeader()
	err = writeCSV(&indexHeader, path.Join(outPath, indexHeaderCSV+".csv"))
	if err != nil {
		log.Error(err)
	}
	indexEntries := databases.GetIndexOffsets()
	err = writeCSV(&indexEntries, path.Join(outPath, indexCSV+".csv"))
	if err != nil {
		log.Error(err)
	}
	indexEntries = databases.GetIndexTags()
	err = writeCSV(&indexEntries, path.Join(outPath, indexTagsCSV+".csv"))
	if err != nil {
		log.Error(err)
	}
	for _, s := range databases.GetDatabases() {
		err = writeCSV(databases.GetEntries(s), path.Join(outPath, fmt.Sprintf("%s.csv", s)))
		if err != nil {
			log.Error(err)
		}
	}
}

func writeCSV(t interface{}, filename string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}

	err = gocsv.MarshalFile(t, f)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}
