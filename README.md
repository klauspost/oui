# OUI Library
Library and microservice for looking up manufacturers from MAC addresses.

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

## Using the library

This shows the basic usage if the library. Opening a database and querying.

```Go
import "github.com/klauspost/oui"

func main() {
    // Load the content from "oui.txt" into a static database
	db, err = oui.OpenStaticFile("oui.txt")

	// Query a mac address
	entry, err := db.Query("D0-DF-9A-D8-44-4B")

	// If error is nil, we have a result in "entry"
}
```
Note hat only the `D0-DF-9A` part of the MAC address is used. The parser is flexible, and will allow colons instead of dashes, or even no separator at all, so these strings will return the same results: `D0-DF-9A`, `D0:DF:9A` & `D0DF9A`. The only thing to note is that you cannot omit zeros, so `00-00-00` must be fully filled.

When you initially load the database, you can specify that you want to be able to update it. Therefore this is safe:
```Go
import "github.com/klauspost/oui"

func main() {
	// Load the database from "oui.txt" into an updateable database.
	db, err = oui.OpenFile("oui.txt")

	go func() {
		for {
			// It is safe to keep querying the database while it is being updated.
			entry, err := db.Query("D0-DF-9A-D8-44-4B")
		}
	}()
	
	// Update the database while it is being queried.
	UpdateFile(db, "oui-newer.txt")
}
```

There are several advanced features, that allow you to specify the MAC address as byte values, and you can even request the raw static database for even faster lookups if you are doing many millions lookups per second. See the [Godoc reference](https://godoc.org/github.com/klauspost/oui) for more information on this.

## Using the server
### Downloading and build:

This method requires a [Go installation](https://golang.org/doc/install).

```
go get github.com/klauspost/oui/ouiserver/...
go install github.com/klauspost/oui/ouiserver
```

This should build a "ouiserver" executable in your gopath.

###Service Options
```
Usage of ouiserver:
  -listen=":5000": Listen address and port, for instance 127.0.0.1:5000
  -open="oui.txt": File name with oui.txt to open. Set to 'http' to download
  -origin="*": Value sent in the "Access-Control-Allow-Origin" header.
  -pretty: Will output be formatted with newlines and intentation
  -threads=4: Number of threads to use. Defaults to number of detected cores
  -update-every="": Duration between reloading the database as 'cronexpr'. 
                    Examples are '@daily', '@weekly', '@monthly'
```
The `open` parameter accepts files or a http URL. If you specify `http`, the server will attempt to download the latest version from [IEEE](http://standards-oui.ieee.org/oui.txt).

The `update-every` expression is a 'cronexpr', that allow you to precisely give update intervals. For more information on the syntax, see the [Golang Cron expression parser](https://github.com/gorhill/cronexpr) documentation.

### Querying the Server

Once the service is running, point your browser to ```http://localhost:5000/D0-DF-9A-D8-44-4B```. You can replace "D0-DF-9A-D8-44-4B" with the Mac Address you would like to look up. You can also specify the MAC address as a parameter named "mac".

Note hat only the `D0-DF-9A` part of the MAC address is used. The parser is flexible, and will allow colons instead of dashes, or even no separator at all, so these strings will return the same results: `D0-DF-9A`, `D0:DF:9A` & `D0DF9A`. The only thing to note is that you cannot omit zeros, so `00-00-00` must be fully filled.

Currently looking up the address above yields:
```json
{
  "data": {
    "manufacturer": "Liteon Technology Corporation",
    "address": [
      "Taipei  23585",
      "TAIWAN, PROVINCE OF CHINA"
    ],
    "prefix": "d0:df:9a",
    "country": "TAIWAN, PROVINCE OF CHINA"
  }
}
```
If you query a OUI that doesn't exist in the database, you will get a returncode 404 with this message:
```json
{
  "error": "not found in db"
}
```

If any error occurs you will get a status 400 with an error message, for instance:
```json
{
  "error": "invalid mac address '54-CD-': Address element 3 () is not 2 characters"
}
```
The time specified in the database as the generation time is sent as "Last-Modified" header. 

## Appengine

A special version of the server has been built for app-engine. It can be found in the `appengine` folder.

This version operates entirely from memory, and updates itself every 24 hours. To see a live running version of it you can go here: http://mac-oui.appspot.com/00-00-00


# License

This code is published under an MIT license. See LICENSE file for more information.
