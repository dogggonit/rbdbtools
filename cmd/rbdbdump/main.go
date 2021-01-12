package main

import (
	"flag"
	"fmt"
	"path"
	"rbdbtools/internal/app/rbdbdump"
	"rbdbtools/pkg/logger"
	"rbdbtools/tools"
)

func main() {
	in := flag.String("in", "./.rockbox/", "directory containing database files")
	out := flag.String("out", "./csv/", "directory to output database dumps to (will be created if not exists)")
	csv := flag.Bool("csv", false, "save as csv instead of xlsx")
	flag.Parse()

	missingDBs := make([]string, 0)
	for i := 0; i < 9; i++ {
		if db := fmt.Sprintf("database_%d.tcd", i); !tools.FileExists(path.Join(*in, db)) {
			missingDBs = append(missingDBs, db)
		}
	}
	if idx := "database_idx.tcd"; !tools.FileExists(path.Join(*in, idx)) {
		missingDBs = append(missingDBs, idx)
	}

	if len(missingDBs) > 0 {
		dbs := ""
		for i, e := range missingDBs {
			dbs += e
			if i < len(missingDBs)-1 {
				dbs += ", "
			}
		}

		log := logger.New()
		log.Fatalf("Cannot find databases: %s", dbs)
	} else {
		rbdbdump.Rbdbdump(*in, *out, *csv)
	}
}
