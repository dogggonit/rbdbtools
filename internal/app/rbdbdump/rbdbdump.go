package rbdbdump

import (
	"log"
	"rbdbtools/pkg/database"
	"rbdbtools/pkg/xlsxdb"
)

func Rbdbdump(dbPath string, outPath string, toCsv bool) {
	dbFiles, err := database.GetTagCacheFiles(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	db, err := dbFiles.ToDatabase()
	if err != nil {
		log.Fatal(err)
	}

	xlsx, err := xlsxdb.New(db)
	if err != nil {
		log.Fatal(err)
	}

	if toCsv {
		err = xlsx.WriteCSV(outPath)
	} else {
		err = xlsx.WriteXlsx(outPath)
	}

	if err != nil {
		log.Fatal(err)
	}
}
