// +build ignore

package main

// To run, execute: go run updatedb.go

import (
	"fmt"
	"github.com/klauspost/oui"
)

// Example of opening a database and updating it with new content.
func main() {
	// Load the content from "sampledb.txt" into a dynamic database
	db, err := oui.OpenFile("sampledb.txt")
	if err != nil {
		panic(err)
	}

	// Get an entry
	entry, err := db.Query("00-60-94")
	if err != nil {
		panic(err)
	}
	fmt.Println("\n*** First Lookup:\n" + entry.String())

	// Update the database
	err = oui.UpdateFile(db, "sampledb2.txt")
	if err != nil {
		panic(err)
	}

	// Get an entry
	entry, err = db.Query("00-60-94")
	if err != nil {
		panic(err)
	}
	fmt.Println("\n*** Second Lookup:\n" + entry.String())
}
