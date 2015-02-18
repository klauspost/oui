// Oui Lookup Library
//
// This package allows lookin up manufacturer information from
// MAC addresses.
//
// Package home: http://github.com/klauspost/oui
//
package oui

/**
 * The MIT License (MIT)

 * Copyright (c) 2015 Klaus Post

 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 */

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Internal representation of the database content
type ouiDB map[[3]byte]Entry

// Set an element to contain this value
func (db ouiDB) set(hw HardwareAddr, e Entry) {
	db[[3]byte(hw)] = e
}

// Delete an element. If the element does not exist,
// the function will just return.
func (db ouiDB) del(hw HardwareAddr) {
	delete(db, [3]byte(hw))
}

// This interface can be used to access the raw
// database. This interface is available on Static databases.
type RawGetter interface {
	RawDB() map[[3]byte]Entry
}

// ErrNotFound will be returned when LookUp function fails
// to find the entry in the database.
var ErrNotFound = errors.New("not found")

// OuiDB represents a database that allow you to look up Hardware Addresses
type OuiDB interface {
	// Query the database for an entry based on the mac address
	// If none are found ErrNotFound will be returned.
	Query(string) (*Entry, error)

	// Look up a hardware address and return the entry if any are found.
	// If none are found ErrNotFound will be returned.
	LookUp(HardwareAddr) (*Entry, error)

	// Returns the generation time of the database
	// May return the zero time if unparsable
	Generated() time.Time

	// Internal functions
	set(HardwareAddr, Entry)
	generatedAt(*time.Time)
}

// StaticDB is a database containing OUI entries that doesn't
// allow updates, but on the other hand allows access to the
// underlying data structure.
// See the OuiDB interface for query functions.
type StaticDB interface {
	OuiDB
	RawGetter
}

// DynamicDB is a database containing OUI entries that
// can be safely updated while queries are running.
// See the OuiDB interface for query functions.
type DynamicDB interface {
	OuiDB
	Updater
}

// Create a new dynamic database with optional content.
// You can pass nil as parameter, which will initialize the database.
// A database returned from this can be expected to implement the Updater interface.
func newDynamic(c map[[3]byte]Entry) DynamicDB {
	if c == nil {
		c = make(map[[3]byte]Entry)
	}
	return &updateableDB{ouiDB: c}
}

// Create a new static database with optional content.
// You can pass nil as parameter, which will initialize the database.
// A database returned from this can be expected to implement the RawGetter interface.
func newStatic(c map[[3]byte]Entry) StaticDB {
	if c == nil {
		c = make(map[[3]byte]Entry)
	}
	return &staticDB{ouiDB: c}
}

// A static database
type staticDB struct {
	ouiDB
	dbTime time.Time
}

// Check we implement the interfaces we promise
var _ OuiDB = &staticDB{}
var _ RawGetter = staticDB{}

// Satisfy the RawGetter interface
func (db staticDB) RawDB() map[[3]byte]Entry {
	return db.ouiDB
}

// Query the database for an entry based on the mac address
// If none are found ErrNotFound will be returned.
func (db staticDB) Query(mac string) (*Entry, error) {
	hw, err := ParseMac(mac)
	if err != nil {
		return nil, err
	}
	return db.LookUp(*hw)
}

// LookUp a hardware address and return the entry if any are found.
// If none are found ErrNotFound will be returned.
func (o staticDB) LookUp(hw HardwareAddr) (*Entry, error) {
	e, ok := o.ouiDB[hw]
	if !ok {
		return nil, ErrNotFound
	}
	return &e, nil
}

// Get the generated time
func (o staticDB) Generated() time.Time {
	return time.Time(o.dbTime)
}

// Update "generated at" time
func (d *staticDB) generatedAt(t *time.Time) {
	if t == nil {
		return
	}
	d.dbTime = *t
}

// An updateable database.
// There is a mutex protecting read/write access to the database.
type updateableDB struct {
	ouiDB
	dbTime time.Time
	mu     sync.RWMutex
}

// Check we implement the interfaces we promise
var _ Updater = &updateableDB{}
var _ OuiDB = &updateableDB{}

// Query the database for an entry based on the mac address
// If none are found ErrNotFound will be returned.
func (db *updateableDB) Query(mac string) (*Entry, error) {
	hw, err := ParseMac(mac)
	if err != nil {
		return nil, err
	}
	return db.LookUp(*hw)
}

// Look up a hardware address and return the entry if any are found.
// If none are found ErrNotFound will be returned.
func (o *updateableDB) LookUp(hw HardwareAddr) (*Entry, error) {
	o.mu.RLock()
	e, ok := o.ouiDB[hw]
	o.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	return &e, nil
}

// Get the generated time
func (o *updateableDB) Generated() time.Time {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.dbTime
}

// Update "generated at" time
func (o *updateableDB) generatedAt(t *time.Time) {
	if t == nil {
		return
	}
	o.mu.Lock()
	o.dbTime = *t
	o.mu.Unlock()
}

// Update the database and replace content with the supplied content.
func (o *updateableDB) updateDb(db ouiDB, t *time.Time) {
	o.mu.Lock()
	o.ouiDB = db
	o.generatedAt(t)
	o.mu.Unlock()
}

// UpdateEntry will update/add a single entry to the database.
func (o *updateableDB) UpdateEntry(hw HardwareAddr, e Entry) {
	o.mu.Lock()
	o.ouiDB.set(hw, e)
	o.mu.Unlock()
}

// DeleteEntry will remove an entry from the database.
// If the element does not exist, the function will just return.
func (o *updateableDB) DeleteEntry(hw HardwareAddr) {
	o.mu.Lock()
	o.ouiDB.del(hw)
	o.mu.Unlock()
}

// The Updater interface will be satisfied if the database was opened as a dynamic database.
// This can be used to safely update the database, even while queries are running.
type Updater interface {
	// UpdateEntry will update/add a single entry to the database.
	UpdateEntry(HardwareAddr, Entry)

	// DeleteEntry will remove an entry from the database. If the element does not exist, nothing should happen
	DeleteEntry(HardwareAddr)

	updateDb(ouiDB, *time.Time)
}

// Read an oui file.
func scanOUI(in io.Reader, db ouiDB) (*time.Time, error) {
	buffered := bufio.NewReader(in)
	scanner := bufio.NewScanner(buffered)
	re := regexp.MustCompile(`((?:(?:[0-9a-zA-Z]{2})[-:]){2,5}(?:[0-9a-zA-Z]{2}))(?:/(\w{1,2}))?`)
	var generated *time.Time

	for scanner.Scan() {
		if len(scanner.Text()) == 0 || scanner.Text()[0] == '#' {
			continue
		}

		arr := strings.Split(scanner.Text(), "\t")
		if len(arr) == 0 {
			continue
		}
		// Attempt to find generation time
		t0 := strings.TrimSpace(arr[0])
		if strings.HasPrefix(t0, "Generated: ") {
			t0 = t0[11:]
			t, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", t0)
			// We ignore the error
			if err == nil {
				generated = &t
			}
			continue

		}
		matches := re.FindAllStringSubmatch(arr[0], -1)
		if len(matches) == 0 {
			continue
		}

		s := matches[0][1]

		bt, err := ParseMac(s)
		if err != nil {
			return generated, err
		}

		e := Entry{Prefix: *bt, Manufacturer: arr[len(arr)-1]}
		for scanner.Scan() {
			text := scanner.Text()
			if len(text) < 2 {
				break
			}
			if text[0] != '\t' {
				continue
			}
			e.Address = append(e.Address, strings.Trim(text, "\t \r\n"))
		}
		if len(e.Address) > 0 {
			e.Country = e.Address[len(e.Address)-1]
		}

		i := int(bt[0])<<16 | int(bt[1])<<8 | int(bt[2])
		if i&local != 0 {
			e.Local = true
		}
		if i&multicast != 0 {
			e.Multicast = true
		}
		db[*bt] = e
	}
	return generated, nil
}

const local = 0x020000
const multicast = 0x010000

// OpenStatic will read the content of the given reader and return a database with the content.
// You will not be able to update this database, but you can request the raw database
// with the RawDB() function.
func OpenStatic(in io.Reader) (StaticDB, error) {
	dst := make(map[[3]byte]Entry)
	db := newStatic(dst)
	t, err := scanOUI(in, ouiDB(dst))
	db.generatedAt(t)
	return db, err
}

// OpenStaticFile will read the content of a oui.txt file and return a database with the content.
// You will not be able to update this database, but you can request the raw database
// with the RawDB() function.
func OpenStaticFile(name string) (StaticDB, error) {
	dst := make(map[[3]byte]Entry)
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	db := newStatic(dst)
	t, err := scanOUI(file, ouiDB(dst))
	db.generatedAt(t)
	return db, err
}

// OpenStaticHttp will request the content of the URL given, parse it as a oui.txt file
// and return a database with the content.
// You will not be able to update this database, but you can request the raw database
// with the RawDB() function.
func OpenStaticHttp(url string) (StaticDB, error) {
	dst := make(ouiDB)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	db := newStatic(dst)
	t, err := scanOUI(resp.Body, dst)
	db.generatedAt(t)
	return db, err
}

// Open will read the content of the given reader and return a database with the content.
// You can update the returned database using the Update/UpdateFile/UpdateHttp functions.
func Open(in io.Reader) (DynamicDB, error) {
	dst := make(map[[3]byte]Entry)
	db := newDynamic(dst)
	t, err := scanOUI(in, ouiDB(dst))
	db.generatedAt(t)
	return db, err
}

// OpenFile will read the content of a oui.txt file and return a database with the content.
// You can update the returned database using the Update/UpdateFile/UpdateHttp functions.
func OpenFile(name string) (DynamicDB, error) {
	dst := make(map[[3]byte]Entry)
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	db := newDynamic(dst)
	t, err := scanOUI(file, ouiDB(dst))
	db.generatedAt(t)
	return db, err
}

// OpenHttp will request the content of the URL given, parse it as a oui.txt file
// and return a database with the content.
// You can update the returned database using the Update/UpdateFile/UpdateHttp functions.
func OpenHttp(url string) (DynamicDB, error) {
	dst := make(ouiDB)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	db := newDynamic(dst)
	t, err := scanOUI(resp.Body, dst)
	db.generatedAt(t)
	return db, err
}

// Update will read and replace the content of the database.
// The database will remain usable while the update/parsing
// is taking place.
// If an error occurs during read or parsing, the database will not be replaced
// and the previous version will continue to be served.
func Update(db DynamicDB, r io.Reader) error {
	dst := make(ouiDB)
	t, err := scanOUI(r, dst)
	if err != nil {
		return err
	}
	db.updateDb(dst, t)
	return nil
}

// UpdateFile will read a file and replace the content of the database.
// The database will remain usable while the update/parsing
// is taking place.
// If an error occurs during read or parsing, the database will not be replaced
// and the previous version will continue to be served.
func UpdateFile(db DynamicDB, name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	dst := make(ouiDB)
	t, err := scanOUI(file, dst)
	if err != nil {
		return err
	}
	db.updateDb(dst, t)
	return nil
}

// UpdateHttp will download from a URL and replace the content of the database.
// The database will remain usable while the updating/parsing
// is taking place.
// If an error occurs during read or parsing, the database will not be replaced
// and the previous version will continue to be served.
func UpdateHttp(db DynamicDB, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dst := make(ouiDB)
	t, err := scanOUI(resp.Body, dst)
	if err != nil {
		return err
	}
	db.updateDb(dst, t)
	return nil
}

// PrintDb the entire database to stdout.
func PrintDb(db OuiDB) {
	var hw HardwareAddr
	c := 0
	t := time.Now()
	for i := 0; i < 1<<24; i++ {
		hw[2] = byte(i & 0xff)
		hw[1] = byte((i >> 8) & 0xff)
		hw[0] = byte((i >> 16) & 0xff)
		e, err := db.LookUp(hw)
		if err == nil {
			fmt.Printf("%s\n\n", e.String())
			c++
		}
	}
	fmt.Printf("Finished reading %d entries in %v. %d Entries.\n", 1<<24, time.Now().Sub(t), c)
}
