package gocsi

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

const maxuint32 = 4294967295

// ParseVersion parses a string for a CSI version.
func ParseVersion(s string) (csi.Version, bool) {
	if versions := ParseVersions(s); len(versions) > 0 {
		return versions[0], true
	}
	return csi.Version{}, false
}

// ParseVersions parses a string for one or more CSI versions.
func ParseVersions(s string) []csi.Version {
	if s == "" {
		return nil
	}

	rx := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	matches := rx.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil
	}

	versions := make([]csi.Version, len(matches))
	for i, m := range matches {
		major, _ := strconv.Atoi(m[1])
		minor, _ := strconv.Atoi(m[2])
		patch, _ := strconv.Atoi(m[3])
		versions[i].Major = uint32(major)
		versions[i].Minor = uint32(minor)
		versions[i].Patch = uint32(patch)
	}

	return versions
}

// SprintfVersion formats a Version as a string.
func SprintfVersion(v csi.Version) string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// FprintfVersion formats a Version as a string to the specified writer.
func FprintfVersion(w io.Writer, v csi.Version) (int, error) {
	return fmt.Fprintf(w, "%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// CompareVersions compares two versions and returns:
//
//   -1 if a > b
//    0 if a = b
//    1 if a < b
func CompareVersions(a, b *csi.Version) int8 {
	if a == nil && b == nil {
		return 0
	}
	if a != nil && b == nil {
		return -1
	}
	if a == nil && b != nil {
		return 1
	}
	if a.Major > b.Major {
		return -1
	}
	if a.Major < b.Major {
		return 1
	}
	if a.Minor > b.Minor {
		return -1
	}
	if a.Minor < b.Minor {
		return 1
	}
	if a.Patch > b.Patch {
		return -1
	}
	if a.Patch < b.Patch {
		return 1
	}
	return 0
}

// GetCSIEndpoint returns the network address specified by the
// environment variable CSI_ENDPOINT.
func GetCSIEndpoint() (network, addr string, err error) {
	protoAddr := os.Getenv(CSIEndpoint)
	if emptyRX.MatchString(protoAddr) {
		return "", "", ErrMissingCSIEndpoint
	}
	return ParseProtoAddr(protoAddr)
}

// GetCSIEndpointListener returns the net.Listener for the endpoint
// specified by the environment variable CSI_ENDPOINT.
func GetCSIEndpointListener() (net.Listener, error) {
	proto, addr, err := GetCSIEndpoint()
	if err != nil {
		return nil, err
	}
	return net.Listen(proto, addr)
}

const (
	protoAddrGuessPatt = `(?i)^(?:tcp|udp|ip|unix)[^:]*://`

	protoAddrExactPatt = `(?i)^((?:(?:tcp|udp|ip)[46]?)|` +
		`(?:unix(?:gram|packet)?))://(.+)$`
)

var (
	emptyRX          = regexp.MustCompile(`^\s*$`)
	protoAddrGuessRX = regexp.MustCompile(protoAddrGuessPatt)
	protoAddrExactRX = regexp.MustCompile(protoAddrExactPatt)
)

// ErrParseProtoAddrRequired occurs when an empty string is provided
// to ParseProtoAddr.
var ErrParseProtoAddrRequired = errors.New(
	"non-empty network address is required")

// ParseProtoAddr parses a Golang network address.
func ParseProtoAddr(protoAddr string) (proto string, addr string, err error) {

	if emptyRX.MatchString(protoAddr) {
		return "", "", ErrParseProtoAddrRequired
	}

	// If the provided network address does not begin with one
	// of the valid network protocols then treat the string as a
	// file path.
	//
	// First check to see if the file exists at the specified path.
	// If it does then assume it's a valid file path and return it.
	//
	// Otherwise attempt to create the file. If the file can be created
	// without error then remove the file and return the result a UNIX
	// socket file path.
	if !protoAddrGuessRX.MatchString(protoAddr) {

		// If the file already exists then assume it's a valid sock
		// file and return it.
		if _, err := os.Stat(protoAddr); !os.IsNotExist(err) {
			return "unix", protoAddr, nil
		}

		f, err := os.Create(protoAddr)
		if err != nil {
			return "", "", fmt.Errorf(
				"invalid implied sock file: %s: %v", protoAddr, err)
		}
		if err := f.Close(); err != nil {
			return "", "", fmt.Errorf(
				"failed to verify network address as sock file: %s", protoAddr)
		}
		if err := os.RemoveAll(protoAddr); err != nil {
			return "", "", fmt.Errorf(
				"failed to remove verified sock file: %s", protoAddr)
		}
		return "unix", protoAddr, nil
	}

	// Parse the provided network address into the protocol and address parts.
	m := protoAddrExactRX.FindStringSubmatch(protoAddr)
	if m == nil {
		return "", "", fmt.Errorf("invalid network address: %s", protoAddr)
	}
	return m[1], m[2], nil
}

// ParseMap parses a string into a map. The string's expected pattern is:
//
//         KEY1=VAL1, "KEY2=VAL2 ", "KEY 3= VAL3"
//
// The key/value pairs are separated by a comma and optional whitespace.
// Please see the encoding/csv package (https://goo.gl/1j1xb9) for information
// on how to quote keys and/or values to include leading and trailing
// whitespace.
func ParseMap(line string) map[string]string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	r := csv.NewReader(strings.NewReader(line))
	r.TrimLeadingSpace = true

	record, err := r.Read()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		panic(err)
	}

	data := map[string]string{}
	for i := range record {
		p := strings.SplitN(record[i], "=", 2)
		if len(p) == 0 {
			continue
		}
		k := p[0]
		var v string
		if len(p) > 1 {
			v = p[1]
		}
		data[k] = v
	}

	return data
}

// ParseSlice parses a string into a slice. The string's expected pattern is:
//
//         VAL1, "VAL2 ", " VAL3 "
//
// The values are separated by a comma and optional whitespace. Please see
// the encoding/csv package (https://goo.gl/1j1xb9) for information on how to
// quote values to include leading and trailing whitespace.
func ParseSlice(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	r := csv.NewReader(strings.NewReader(line))
	r.TrimLeadingSpace = true

	record, err := r.Read()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		panic(err)
	}

	return record
}

// ParseMapWS parses a string into a map. The string's expected pattern is:
//
//         KEY1=VAL1 KEY2="VAL2 " "KEY 3"=' VAL3'
//
// The key/value pairs are separated by one or more whitespace characters.
// Keys and/or values with whitespace should be quoted with either single
// or double quotes.
func ParseMapWS(line string) map[string]string {
	if line == "" {
		return nil
	}

	var (
		escp bool
		quot rune
		ckey string
		keyb = &bytes.Buffer{}
		valb = &bytes.Buffer{}
		word = keyb
		data = map[string]string{}
	)

	for i, c := range line {
		// Check to see if the character is a quote character.
		switch c {
		case '\\':
			// If not already escaped then activate escape.
			if !escp {
				escp = true
				continue
			}
		case '\'', '"':
			// If the quote or double quote is the first char or
			// an unescaped char then determine if this is the
			// beginning of a quote or end of one.
			if i == 0 || !escp {
				if quot == c {
					quot = 0
				} else {
					quot = c
				}
				continue
			}
		case '=':
			// If the word buffer is currently the key buffer,
			// quoting is not enabled, and the preceeding character
			// is not the escape character then the equal sign indicates
			// a transition from key to value.
			if word == keyb && quot == 0 && !escp {
				ckey = keyb.String()
				keyb.Reset()
				word = valb
				continue
			}
		case ' ', '\t':
			// If quoting is not enabled and the preceeding character is
			// not the escape character then record the value into the
			// map and fast-forward the cursor to the next, non-whitespace
			// character.
			if quot == 0 && !escp {
				// Record the value into the map for the current key.
				if ckey != "" {
					data[ckey] = valb.String()
					valb.Reset()
					word = keyb
					ckey = ""
				}
				continue
			}
		}
		if escp {
			escp = false
		}
		word.WriteRune(c)
	}

	// If the current key string is not empty then record it with the value
	// buffer's string value as a new pair.
	if ckey != "" {
		data[ckey] = valb.String()
	}

	return data
}
