package rbdbdump

import (
	"os"
	"rbdbtools/tools"
)

func Rbdbdump(dbPath string, outPath string, toCsv bool) {
	if !tools.DirExists(outPath) {
		err := os.MkdirAll(outPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	if toCsv {
		csv(dbPath, outPath)
	} else {
		toXlsx(dbPath, outPath)
	}
}
