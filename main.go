package main

import (
	"github.com/jucardi/go-logger-lib/log"
	"github.com/jucardi/go-osx/paths"
	"github.com/spf13/cobra"
	"path/filepath"
)

var (
	rootCmd = &cobra.Command{
		Use:   "protoc-go-inject-tag",
		Short: "Injects Tags to generated protobuf go files.",
		Run:   run,
	}
	cleanup = false
)

func main() {
	rootCmd.Flags().StringP("input", "i", "", "The input file or wildcard that matches files to process")
	rootCmd.Flags().StringArrayP("XXXSkip", "x", []string{}, "The tags to also mark as ignored for XXX fields, Eg: if using 'yaml' as this value, the tags section will result in: `json:\"-\" yaml:\"-\"`")
	rootCmd.Flags().BoolP("cleanup", "c", false, "Cleans up any comment lines with the '\\ @inject_tag' from the autogenerated code when done.")
	rootCmd.Flags().BoolP("verbose", "v", false, "Enables debug mode")

	log.PanicErr(
		rootCmd.Execute(),
	)
}

func run(cmd *cobra.Command, _ []string) {
	inputFile, _ := cmd.Flags().GetString("input")
	xxxTags, _ := cmd.Flags().GetStringArray("XXXSkip")
	verbose, _ := cmd.Flags().GetBool("verbose")
	cleanup, _ = cmd.Flags().GetBool("cleanup")

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	if inputFile == "" {
		log.Fatal("The input cannot be empty", cmd.Usage())
	}

	if exists, err := paths.Exists(inputFile); err != nil {
		log.FatalErr(err)
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

func processFile(inputFile string, xxxTags []string) {
	if len(inputFile) == 0 {
		log.Fatal("input file is mandatory")
	}

	areas, err := parseFile(inputFile, xxxTags)
	log.FatalErr(err)
	log.FatalErr(
		writeFile(inputFile, areas),
	)
}
