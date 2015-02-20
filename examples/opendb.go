// +build ignore

package main

// To run, execute: go run opendb.go

import (
	"github.com/klauspost/oui"
)

// Example of opening a database
func main() {
	// Load the content from "sampledb.txt" into a static database
	db, err := oui.OpenStaticFile("sampledb.txt")
	if err != nil {
		panic(err)
	}

	oui.PrintDb(db)
}
