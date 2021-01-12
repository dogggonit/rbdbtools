package main

import (
	"errors"
	"flag"
	"rbdbtools/internal/app/rbdbgen"
	"rbdbtools/pkg/logger"
	"rbdbtools/tools"
)

func main() {
	big := flag.Bool("big", false, "use big endian database (coldfire and SH1)")
	t := flag.String("target", "./database/", "directory to output database files to (will be created if not exists)")
	i := flag.String("internal", "", "location of music on internal media")
	e := flag.String("external", "", "location of music on external media")
	flag.Parse()

	log := logger.New()
	if *i == "" && *e == "" {
		log.Fatal(errors.New("internal and/or external must be specified"))
	} else if *i != "" && !tools.DirExists(*i) {
		log.Fatal("internal directory does not exist")
	} else if *e != "" && !tools.DirExists(*e) {
		log.Fatal("external directory does not exist")
	} else {
		rbdbgen.Rbdbgen(*big, *t, *i, *e)
	}
}
