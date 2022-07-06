package main

import (
	"flag"
	"os"
)

var (
	flagBaseResumeFile   string
	flagSecretResumeFile string
	flagControlsFile     string
	flagGeneratedPdf     string
	flagShowHelp         bool
)

func initFlags() {
	flag.StringVar(&flagBaseResumeFile, "base-resume", "conf/resume/base.yaml", "Path to base resume file to use")
	flag.StringVar(&flagSecretResumeFile, "secret-resume", "conf/resume/secret.yaml", "Path to secret resume file to use")
	flag.StringVar(&flagControlsFile, "controls", "conf/controls/default.yaml", "Path to the controls file to use")
	flag.StringVar(&flagGeneratedPdf, "output-pdf", "", "The filename to use for the generated PDF")
	flag.BoolVar(&flagShowHelp, "help", false, "Show help")
}

func parseFlags() {
	flag.Parse()

	if flagShowHelp {
		flag.Usage()

		os.Exit(0)
	}
}
