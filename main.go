package main

import (
	"flag"
	"github.com/jucardi/go-logger-lib/log"
	"github.com/jucardi/go-osx/paths"
	"path/filepath"
	"strings"
)

func main() {
	var inputFile string
	var xxxTags string
	flag.StringVar(&inputFile, "input", "", "path to input file")
	flag.StringVar(&xxxTags, "XXX_skip", "", "skip tags to inject on XXX fields")

	flag.Parse()

	if exists, err := paths.Exists(inputFile); err != nil {
		log.Fatal(err)
	} else if exists {
		processFile(inputFile, xxxTags)
		return
	}

	files, err := filepath.Glob(inputFile)
	log.FatalErr(err)

	for _, file := range files {
		processFile(file, xxxTags)
	}
}

func processFile(inputFile, xxxTags string) {
	var xxxSkipSlice []string
	if len(xxxTags) > 0 {
		xxxSkipSlice = strings.Split(xxxTags, ",")
	}

	if len(inputFile) == 0 {
		log.Fatal("input file is mandatory")
	}

	areas, err := parseFile(inputFile, xxxSkipSlice)
	if err != nil {
		log.Fatal(err)
	}
	if err = writeFile(inputFile, areas); err != nil {
		log.Fatal(err)
	}
}
