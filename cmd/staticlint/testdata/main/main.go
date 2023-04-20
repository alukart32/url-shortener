package main

import "os"

func notMain() {
	os.Exit(0)
	other()
}

func main() {
	os.Exit(1) // want "os.Exit call"
	notMain()
}
