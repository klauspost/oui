package main

import (
	"encoding/json"
	"flag"
	"github.com/gorhill/cronexpr"
	"github.com/klauspost/oui"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strings"
	"time"
)

var ouiFile = flag.String("open", "oui.txt", "File name with oui.txt to open. Set to 'http' to download")
var listen = flag.String("listen", ":5000", "Listen address and port, for instance 127.0.0.1:5000")
var threads = flag.Int("threads", runtime.NumCPU(), "Number of threads to use. Defaults to number of detected cores")
var pretty = flag.Bool("pretty", false, "Should output be formatted with newlines and intentation")
var update = flag.String("update-every", "", "Duration between reloading the database as 'cronexpr'. Examples are '@weekly', '@monthly'.")

//go:generate: ffjson -nodecoder $(GOFILE)

// ffjson: nodecoder
type Response struct {
	Data  *oui.Entry `json:"data,omitempty"`
	Error string     `json:"error,omitempty"`
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(*threads)

	var cron *cronexpr.Expression
	if *update != "" {
		cron = cronexpr.MustParse(*update)
	}

	var db oui.OuiDb
	url := ""
	fileName := ""
	var err error

	if strings.HasPrefix(*ouiFile, "http") {
		url = *ouiFile
		if url == "http" {
			url = "http://standards-oui.ieee.org/oui.txt"
		}
		log.Println("Downloading new Db from: " + url)
		db, err = oui.OpenHttp(url, cron != nil)
		if err != nil {
			log.Fatalf("Error downloading:%s", err.Error())
		}
	} else {
		fileName = *ouiFile
		log.Println("Opening database from: " + fileName)
		db, err = oui.OpenFile(fileName, cron != nil)
		if err != nil {
			log.Fatalf("Error updating file:%s", err.Error())
		}
	}
	log.Printf("Database generated at %s\n", db.Generated().Local().String())

	// Start updater if needed.
	if cron != nil {
		go func() {
			for {
				// Sleep until next update
				next := cron.Next(time.Now())
				log.Println("Next update: " + next.String())
				time.Sleep(next.Sub(time.Now()))
				if url != "" {
					log.Println("Updating db from: " + url)
					err := oui.UpdateHttp(db, url)
					if err != nil {
						log.Printf("Error downloading update:%s", err.Error())
					} else {
						log.Println("Updated Successfully")
					}
				} else {
					log.Println("Updating db with file: " + fileName)
					err := oui.UpdateFile(db, fileName)
					if err != nil {
						log.Printf("Error loading update:%s", err.Error())
					} else {
						log.Println("Updated Successfully")
					}
				}
			}
		}()
	}

	// We dereference this to avoid a pretty big penalty under heavy load.
	prettyL := *pretty

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		var mac string
		var hw *oui.HardwareAddr

		// Prepare the response and queue sending the result.
		res := &Response{}

		defer func() {
			var j []byte
			var err error
			if prettyL {
				j, err = json.MarshalIndent(res, "", "  ")
			} else {
				j, err = res.MarshalJSON()
			}
			if err != nil {
				log.Fatal(err)
			}
			w.Write(j)
		}()

		mac = req.URL.Query().Get("mac")
		if mac == "" {
			mac = strings.Trim(req.URL.Path, "/")
		}
		hw, err := oui.ParseMac(mac)
		if err != nil {
			res.Error = err.Error()
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		entry, err := db.LookUp(*hw)
		if err != nil {
			if err == oui.ErrNotFound {
				res.Error = "not found in db"
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			res.Error = err.Error()
			return
		}
		res.Data = entry
	})

	log.Println("Listening on " + *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
