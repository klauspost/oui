package oui

import (
	"strings"
)

//go:generate: ffjson -nodecoder $(GOFILE)

// A database Entry the represents the data in the oui database.
// Local and Multicast
type Entry struct {
	Manufacturer string       `json:"manufacturer"`
	Address      []string     `json:"address"`
	Prefix       HardwareAddr `json:"prefix"`
	Country      string       `json:"country,omitempty"`
	Local        bool         `json:"local,omitempty"`
	Multicast    bool         `json:"multicast,omitempty"`
}

// Returns a formatted string representation of the entry
func (e Entry) String() string {
	t := []string{"Prefix: " + e.Prefix.String(), "Manufacturer: " + e.Manufacturer}
	if len(e.Address) > 0 {
		a := strings.Join(e.Address, "\n\t")
		t = append(t, "Address:", "\t"+a)
	}
	if e.Local {
		t = append(t, "* Locally Administered")
	}
	if e.Multicast {
		t = append(t, "* Multicast")
	}
	return strings.Join(t, "\n")
}
