package oui

import (
	"fmt"
	"strings"
)

// HardwareAddr is the OUI part of a mac address.
// An easy way to get a hardware address is to use the ParseMac function
// The hardware address is in transmission bit order.
type HardwareAddr [3]byte

// String returns a hex string identifying the OUI.
// This will be as "xx:yy:zz" where elements are separated by ':'
// and written in transmission bit order
func (h HardwareAddr) String() string {
	return fmt.Sprintf("%02x:%02x:%02x", h[0], h[1], h[2])
}

// This function will return the address as a quoted hex string.
func (h HardwareAddr) MarshalJSON() ([]byte, error) {
	return []byte(`"` + h.String() + `"`), nil
}

// This function will return the address as a quoted hex string.
func (h *HardwareAddr) UnmarshalJSON(in []byte) error {
	n, err := ParseMac(strings.Trim(string(in), `" `))
	if err != nil {
		return err
	}
	*h = *n
	return nil
}

// Local returns true if the address is in the
// "locally administered" segment.
func (h HardwareAddr) Local() bool {
	return (h[0] & 2) > 0
}

// Multicast returns true if the address is in the
// multicast segment.
func (h HardwareAddr) Multicast() bool {
	return (h[0] & 1) > 0
}

// This error will be returned by ParseMac
// if the Mac address cannot be decoded.
type ErrInvalidMac struct {
	Reason string
	Mac    string
}

// Error returns a string representation of the error.
func (e ErrInvalidMac) Error() string {
	return "invalid mac address '" + e.Mac + "': " + e.Reason
}

// ParseMac will parse a string Mac address and return the first 3 entries.
// It will attempt to find a separator, ':' and '-' supported.
// If none of these are matched, it will assume there is none.
func ParseMac(mac string) (*HardwareAddr, error) {
	// Attempt to find a separator, ':' and '-' supported.
	if len(mac) < 6 {
		return nil, ErrInvalidMac{Reason: "Mac address too short. Should be at least 6 characters", Mac: mac}
	}
	var separator *byte
	var s []string

	if mac[2] == ':' || mac[2] == '-' {
		b := mac[2]
		separator = &b
		s = strings.Split(mac, string(*separator))
	} else {
		for i := 0; i < len(mac)-1; i += 2 {
			s = append(s, mac[i:i+2])
		}
	}
	if len(s) < 3 {
		return nil, ErrInvalidMac{Reason: "Unable to find at least 3 address elements", Mac: mac}
	}
	hw := HardwareAddr{}
	for i, p := range s {
		if i >= 3 {
			break
		}
		if len(p) != 2 {
			return nil, ErrInvalidMac{Reason: fmt.Sprintf("Address element %d (%s) is not 2 characters", i+1, p), Mac: mac}
		}
		var b byte
		n, err := fmt.Sscanf(p, "%x", &b)
		if n != 1 {
			return nil, ErrInvalidMac{Reason: fmt.Sprintf("Address element %d (%s) cannot be parsed as hex value: %v", i+1, p, err), Mac: mac}
		}
		hw[i] = b
	}
	return &hw, nil
}
