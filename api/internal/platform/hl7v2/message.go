package hl7v2

import (
	"fmt"
	"strings"
	"time"
)

// Message represents a parsed HL7v2 message.
type Message struct {
	Type         string    // MSH-9 message type (e.g. "ADT^A01")
	ControlID    string    // MSH-10
	Version      string    // MSH-12 (e.g. "2.5.1")
	Timestamp    time.Time // MSH-7
	SendingApp   string    // MSH-3
	SendingFac   string    // MSH-4
	ReceivingApp string    // MSH-5
	ReceivingFac string    // MSH-6
	Segments     []Segment
}

// Segment represents a single HL7v2 segment.
type Segment struct {
	Name   string  // e.g. "MSH", "PID", "OBR", "OBX"
	Fields []Field
}

// Field represents a field which can have components and repetitions.
type Field struct {
	Value      string
	Components []string   // Component-separated (^)
	Repeats    [][]string // Repetition-separated (~), each with components
}

// Parse parses raw HL7v2 message bytes into a structured Message.
// It supports \r, \n, and \r\n line endings for segment separation.
func Parse(raw []byte) (*Message, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("hl7v2: message is empty")
	}

	text := string(raw)

	// Normalize line endings: replace \r\n with \r, then replace \n with \r
	text = strings.ReplaceAll(text, "\r\n", "\r")
	text = strings.ReplaceAll(text, "\n", "\r")

	// Split into segment lines
	lines := strings.Split(text, "\r")

	// Filter empty lines
	var segmentLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			segmentLines = append(segmentLines, line)
		}
	}

	if len(segmentLines) == 0 {
		return nil, fmt.Errorf("hl7v2: no segments found")
	}

	// First segment must be MSH
	if !strings.HasPrefix(segmentLines[0], "MSH") {
		return nil, fmt.Errorf("hl7v2: first segment must be MSH, got %q", segmentLines[0][:min(3, len(segmentLines[0]))])
	}

	msg := &Message{}

	for _, line := range segmentLines {
		seg, err := parseSegment(line)
		if err != nil {
			return nil, fmt.Errorf("hl7v2: failed to parse segment: %w", err)
		}
		msg.Segments = append(msg.Segments, seg)
	}

	// Extract MSH header fields
	if err := msg.extractMSHFields(); err != nil {
		return nil, err
	}

	return msg, nil
}

// parseSegment parses a single segment line into a Segment struct.
func parseSegment(line string) (Segment, error) {
	if len(line) < 3 {
		return Segment{}, fmt.Errorf("segment too short: %q", line)
	}

	seg := Segment{}

	// MSH is special: the field separator (|) is MSH-1 itself.
	if strings.HasPrefix(line, "MSH") {
		seg.Name = "MSH"
		if len(line) < 4 {
			return seg, nil
		}

		fieldSep := string(line[3]) // should be |
		// MSH-2 is the encoding characters (typically ^~\&)
		// For MSH, we split the rest starting after MSH|
		rest := line[4:] // everything after "MSH|"
		parts := strings.Split(rest, fieldSep)

		// MSH-1 = | (the field separator)
		// MSH-2 = first element of parts (encoding characters)
		// MSH-3 = second element of parts, etc.
		// We store fields starting from MSH-1.
		// fields[0] = MSH-1 = "|"
		// fields[1] = MSH-2 = encoding chars
		// fields[2] = MSH-3 = sending app, etc.

		fieldSepField := Field{
			Value:      fieldSep,
			Components: []string{fieldSep},
		}
		seg.Fields = append(seg.Fields, fieldSepField)

		for _, part := range parts {
			seg.Fields = append(seg.Fields, parseField(part))
		}
	} else {
		// Normal segments: name|field1|field2|...
		parts := strings.SplitN(line, "|", 2)
		seg.Name = parts[0]

		if len(parts) > 1 {
			fields := strings.Split(parts[1], "|")
			for _, f := range fields {
				seg.Fields = append(seg.Fields, parseField(f))
			}
		}
	}

	return seg, nil
}

// parseField parses a single field, handling components (^) and repetitions (~).
func parseField(raw string) Field {
	f := Field{
		Value: raw,
	}

	// Parse repetitions first (~ separator)
	reps := strings.Split(raw, "~")
	for _, rep := range reps {
		components := strings.Split(rep, "^")
		f.Repeats = append(f.Repeats, components)
	}

	// Components from the first repetition (or the whole value if no repetitions)
	if len(f.Repeats) > 0 {
		f.Components = f.Repeats[0]
	} else {
		f.Components = strings.Split(raw, "^")
	}

	return f
}

// extractMSHFields extracts commonly used MSH fields into the Message struct.
func (m *Message) extractMSHFields() error {
	msh := m.GetSegment("MSH")
	if msh == nil {
		return fmt.Errorf("hl7v2: MSH segment not found")
	}

	// MSH field indexing:
	// fields[0] = MSH-1 (|)
	// fields[1] = MSH-2 (^~\&)
	// fields[2] = MSH-3 (sending app)
	// fields[3] = MSH-4 (sending fac)
	// fields[4] = MSH-5 (receiving app)
	// fields[5] = MSH-6 (receiving fac)
	// fields[6] = MSH-7 (timestamp)
	// fields[7] = MSH-8 (security)
	// fields[8] = MSH-9 (message type)
	// fields[9] = MSH-10 (control ID)
	// fields[10] = MSH-11 (processing ID)
	// fields[11] = MSH-12 (version)

	m.SendingApp = mshField(msh, 2)
	m.SendingFac = mshField(msh, 3)
	m.ReceivingApp = mshField(msh, 4)
	m.ReceivingFac = mshField(msh, 5)

	// Parse timestamp (MSH-7)
	tsStr := mshField(msh, 6)
	if tsStr != "" {
		t, err := parseHL7Timestamp(tsStr)
		if err == nil {
			m.Timestamp = t
		}
	}

	// Message type (MSH-9) — includes components like ADT^A01
	m.Type = mshField(msh, 8)

	// Control ID (MSH-10)
	m.ControlID = mshField(msh, 9)

	// Version (MSH-12)
	m.Version = mshField(msh, 11)

	return nil
}

// mshField returns the value of an MSH field by its 0-based index into the Fields slice.
// MSH indexing: Fields[0]=MSH-1, Fields[1]=MSH-2, ... Fields[n]=MSH-(n+1).
func mshField(msh *Segment, index int) string {
	if index >= len(msh.Fields) {
		return ""
	}
	return msh.Fields[index].Value
}

// parseHL7Timestamp parses an HL7v2 timestamp string (YYYYMMDDHHmmss or YYYYMMDD).
func parseHL7Timestamp(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	switch {
	case len(s) >= 14:
		return time.Parse("20060102150405", s[:14])
	case len(s) >= 12:
		return time.Parse("200601021504", s[:12])
	case len(s) >= 8:
		return time.Parse("20060102", s[:8])
	default:
		return time.Time{}, fmt.Errorf("hl7v2: unrecognized timestamp format: %q", s)
	}
}

// GetSegment returns the first segment with the given name, or nil if not found.
func (m *Message) GetSegment(name string) *Segment {
	for i := range m.Segments {
		if m.Segments[i].Name == name {
			return &m.Segments[i]
		}
	}
	return nil
}

// GetSegments returns all segments with the given name.
func (m *Message) GetSegments(name string) []Segment {
	var result []Segment
	for _, seg := range m.Segments {
		if seg.Name == name {
			result = append(result, seg)
		}
	}
	return result
}

// GetField returns the value of a field by 1-based index.
// For non-MSH segments, field index 1 corresponds to Fields[0].
// For MSH, MSH-1 is Fields[0] (the field separator).
func (s *Segment) GetField(index int) string {
	if s.Name == "MSH" {
		// MSH fields: Fields[0]=MSH-1, Fields[1]=MSH-2, etc.
		idx := index - 1
		if idx < 0 || idx >= len(s.Fields) {
			return ""
		}
		return s.Fields[idx].Value
	}

	// Non-MSH segments: Fields[0]=field-1, Fields[1]=field-2, etc.
	idx := index - 1
	if idx < 0 || idx >= len(s.Fields) {
		return ""
	}
	return s.Fields[idx].Value
}

// GetComponent returns a component value by 1-based field and component indices.
// For non-MSH segments, field index 1 corresponds to Fields[0].
func (s *Segment) GetComponent(fieldIdx, compIdx int) string {
	var field *Field

	if s.Name == "MSH" {
		idx := fieldIdx - 1
		if idx < 0 || idx >= len(s.Fields) {
			return ""
		}
		field = &s.Fields[idx]
	} else {
		idx := fieldIdx - 1
		if idx < 0 || idx >= len(s.Fields) {
			return ""
		}
		field = &s.Fields[idx]
	}

	ci := compIdx - 1
	if ci < 0 || ci >= len(field.Components) {
		return ""
	}
	return field.Components[ci]
}

// PatientID returns PID-3.1 (the first component of the patient identifier field).
func (m *Message) PatientID() string {
	pid := m.GetSegment("PID")
	if pid == nil {
		return ""
	}
	return pid.GetComponent(3, 1)
}

// PatientName returns the family and given name from PID-5 (family^given).
func (m *Message) PatientName() (family, given string) {
	pid := m.GetSegment("PID")
	if pid == nil {
		return "", ""
	}
	family = pid.GetComponent(5, 1)
	given = pid.GetComponent(5, 2)
	return family, given
}

// DateOfBirth returns PID-7 (date of birth).
func (m *Message) DateOfBirth() string {
	pid := m.GetSegment("PID")
	if pid == nil {
		return ""
	}
	return pid.GetField(7)
}

// Gender returns PID-8 (administrative sex).
func (m *Message) Gender() string {
	pid := m.GetSegment("PID")
	if pid == nil {
		return ""
	}
	return pid.GetField(8)
}
