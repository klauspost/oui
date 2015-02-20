// +build ignore

package main

// To run, execute: go run query.go

import (
	"fmt"
	"github.com/klauspost/oui"
)

// Example of querying a database
func main() {
	// Load the content from "sampledb.txt" into a static database
	db, err := oui.OpenStaticFile("sampledb.txt")
	if err != nil {
		panic(err)
	}

	// Query on text string
	entry, err := db.Query("00-60-93-98-02-01")
	if err == oui.ErrNotFound {
		fmt.Println("Not found")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println(entry.String())
	}
}
