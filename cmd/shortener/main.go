package main

import (
	"fmt"
	"os"

	"github.com/alukart32/shortener-url/internal/shortener"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()
	shortener.Run()
}

func printBuildInfo() {
	if len(buildVersion) == 0 {
		fmt.Fprintln(os.Stdout, "Build version: N/A")
	} else {
		fmt.Fprintln(os.Stdout, "Build version: "+buildVersion)
	}

	if len(buildDate) == 0 {
		fmt.Fprintln(os.Stdout, "Build date: N/A")
	} else {
		fmt.Fprintln(os.Stdout, "Build date: "+buildDate)
	}

	if len(buildCommit) == 0 {
		fmt.Fprintln(os.Stdout, "Build commit: N/A")
	} else {
		fmt.Fprintln(os.Stdout, "Build commit: "+buildCommit)
	}
}
