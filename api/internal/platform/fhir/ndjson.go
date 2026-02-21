package fhir

import (
	"bufio"
	"encoding/json"
	"io"
)

// NDJSONWriter writes resources in NDJSON (Newline Delimited JSON) format.
// Each resource is serialised as a single JSON line followed by a newline
// character, which is the format required by the FHIR Bulk Data Access
// specification.
type NDJSONWriter struct {
	w *bufio.Writer
}

// NewNDJSONWriter creates a new NDJSONWriter that writes to w.
func NewNDJSONWriter(w io.Writer) *NDJSONWriter {
	return &NDJSONWriter{
		w: bufio.NewWriter(w),
	}
}

// WriteResource serialises resource as a single JSON line followed by a
// newline character. The resource can be any value that is marshallable
// by encoding/json (typically a map[string]interface{} or a struct).
func (n *NDJSONWriter) WriteResource(resource interface{}) error {
	data, err := json.Marshal(resource)
	if err != nil {
		return err
	}
	if _, err := n.w.Write(data); err != nil {
		return err
	}
	if err := n.w.WriteByte('\n'); err != nil {
		return err
	}
	return nil
}

// Flush flushes any buffered data to the underlying writer.
func (n *NDJSONWriter) Flush() error {
	return n.w.Flush()
}
