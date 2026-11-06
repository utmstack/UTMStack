package syslog

import (
	"bufio"
	"strings"
	"testing"
)

// ING-F5: a line that is neither syslog's '<' nor octet-counting's leading
// digit must still be read, not rejected — in particular JSON (Suricata EVE
// sends one JSON object per line over this same TCP syslog port).
func TestDetectFramingMethodAcceptsEveryByteExceptOctetCounting(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  FramingMethod
	}{
		{"syslog PRI", "<34>1 2003-10-11 host app - - msg\n", FramingNewline},
		{"JSON object", `{"event_type":"alert"}` + "\n", FramingNewline},
		{"JSON array", "[1,2,3]\n", FramingNewline},
		{"plain text", "hello world\n", FramingNewline},
		{"octet counting", "23 <34>1 2003-10-11 msg", FramingOctetCounting},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.input))
			got, err := detectFramingMethod(reader)
			if err != nil {
				t.Fatalf("detectFramingMethod: unexpected error %v", err)
			}
			if got != tc.want {
				t.Fatalf("framing method = %v, want %v", got, tc.want)
			}
		})
	}
}

// A JSON line must be read whole, the same as any other newline-delimited
// message — it must not be refused the way it used to be before ING-F5.
func TestReadSyslogMessageAcceptsJSONLine(t *testing.T) {
	line := `{"event_type":"alert","src_ip":"1.2.3.4"}`
	reader := bufio.NewReader(strings.NewReader(line + "\n"))

	got, err := readSyslogMessage(reader)
	if err != nil {
		t.Fatalf("readSyslogMessage: unexpected error %v", err)
	}
	if got != line+"\n" {
		t.Fatalf("message = %q, want %q", got, line+"\n")
	}
}

// A connection carrying several unrelated-looking lines (syslog, JSON, plain
// text) back to back must deliver all of them — none of them may abort
// reading the rest of the stream.
func TestReadSyslogMessageSurvivesMixedLinesOnOneConnection(t *testing.T) {
	lines := []string{
		"<34>1 2003-10-11 host app - - msg one\n",
		`{"event_type":"alert"}` + "\n",
		"plain text line\n",
	}
	reader := bufio.NewReader(strings.NewReader(strings.Join(lines, "")))

	for i, want := range lines {
		got, err := readSyslogMessage(reader)
		if err != nil {
			t.Fatalf("line %d: unexpected error %v", i, err)
		}
		if got != want {
			t.Fatalf("line %d = %q, want %q", i, got, want)
		}
	}
}
