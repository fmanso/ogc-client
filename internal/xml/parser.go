// Package xml provides XML parsing utilities for OGC services.
package xml

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

// Decoder wraps xml.Decoder with OGC-specific utilities.
type Decoder struct {
	*xml.Decoder
}

// NewDecoder creates a new XML decoder for the given reader.
func NewDecoder(r io.Reader) *Decoder {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	dec.CharsetReader = charsetReader
	return &Decoder{Decoder: dec}
}

// charsetReader handles character encoding for XML parsing.
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	// Handle common encodings
	switch strings.ToLower(charset) {
	case "iso-8859-1", "latin1":
		return input, nil // Go's xml package handles this
	case "utf-8", "utf8":
		return input, nil
	default:
		// Return as-is and hope for the best
		return input, nil
	}
}

// Encode encodes a value to XML with standard OGC settings.
func Encode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MustEncode encodes a value to XML, panicking on error.
func MustEncode(v interface{}) []byte {
	b, err := Encode(v)
	if err != nil {
		panic(err)
	}
	return b
}

// TrimElementSpace trims whitespace from element content.
func TrimElementSpace(s string) string {
	return strings.TrimSpace(s)
}

// ParseBool parses a boolean from XML attribute/element.
func ParseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}

// BoolToString converts a bool to XML string representation.
func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
