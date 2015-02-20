// +build ignore

package main

// To run, execute: go run querybytes.go

import (
	"fmt"
	"github.com/klauspost/oui"
)

// Example of looking up with byte values
func main() {
	// Load the content from "sampledb.txt" into a static database
	db, err := oui.OpenStaticFile("sampledb.txt")
	if err != nil {
		panic(err)
	}

	// We create a hardware address with the values we would like to look up:
	hw := oui.HardwareAddr{0x00, 0x60, 0x92}

	// Now we look up
	entry, err := db.LookUp(hw)
	if err == oui.ErrNotFound {
		fmt.Println("Not found")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println(entry.String())
	}
}
