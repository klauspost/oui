# OUI Library
Library for looking up manufacturers from MAC addresses.

This library is a in-memory database that allows to look up manufacturer information based on a MAC address. The library is very lightweight, and allows for million of lookups per second. You can add the database by specifying a file or a URL where the data can be downloaded from. The server support non-interuptible updates and can update the database while the server is running.

An Organizationally Unique Identifier (OUI) is a 24-bit number that uniquely identifies a vendor, manufacturer, or other organization globally or worldwide.

These are purchased from the Institute of Electrical and Electronics Engineers, Incorporated (IEEE) Registration Authority by the "assignee" (IEEE term for the vendor, manufacturer, or other organization). They are used as the first portion of derivative identifiers to uniquely identify a particular piece of equipment as Ethernet MAC addresses, Subnetwork Access Protocol protocol identifiers, World Wide Names for Fibre Channel host bus adapters, and other Fibre Channel and Serial Attached SCSI devices.

In MAC addresses, the OUI is combined with a 24-bit number (assigned by the owner or 'assignee' of the OUI) to form the address. The first three octets of the address are the OUI.

[See full article on Wikipedia](http://en.wikipedia.org/wiki/Organizationally_unique_identifier)

Package home: https://github.com/klauspost/oui

App Engine Live Server: http://mac-oui.appspot.com/00-01-02

Download OUI database: http://standards.ieee.org/develop/regauth/oui/public.html

# Documentation
[![GoDoc][1]][2] [![Build Status][3]][4]
[1]: https://godoc.org/github.com/klauspost/oui?status.svg
[2]: https://godoc.org/github.com/klauspost/oui
[3]: https://travis-ci.org/klauspost/oui.svg
[4]: https://travis-ci.org/klauspost/oui

[Godoc reference](https://godoc.org/github.com/klauspost/oui).

# Usage

```Go
import "github.com/klauspost/oui"

func main() {
    // Load the database from "oui.txt", we don't need an updateable database.
	db, err = oui.OpenFile("oui.txt", false)

	// Query a mac address
	entry, err := db.Query("D0-DF-9A-D8-44-4B")

	// If error is nil, we have a result in "entry"
}
```

# License

This code is published under an MIT license. See LICENSE file for more information.
