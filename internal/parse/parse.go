// Package parse detects the log format and extracts a small structured
// summary. Detection heuristics and summary fields are mirrored by the
// TypeScript implementation in web/src/lib/parser — keep them in sync.
package parse

import (
	"regexp"
	"strconv"
	"strings"
)

type Format string

const (
	FormatFreeRADIUS Format = "freeradius"
	FormatNPS        Format = "nps"
	FormatUnknown    Format = "unknown"
)

// npsRe covers Event Viewer text (6272/6273/6274), the Event Viewer XML view,
// NPS System-log events 13/18, and the IAS/DTS accounting log-file formats.
// `IAS` is word-bounded: unanchored it matched "tobias" or "alias" inside
// FreeRADIUS output and flipped the format to NPS.
var (
	npsRe = regexp.MustCompile(`(?i)Network Policy Server|Reason Code:|Event ID:\s*627[234]|<EventID>627[234]</EventID>|Name="ReasonCode"|\bIAS\b|<Reason-Code data_type=|<Packet-Type data_type=|invalid RADIUS client IP address|Message-Authenticator attribute that is not valid`)
	frRe  = regexp.MustCompile(`(?i)FreeRADIUS|radiusd|Access-(Request|Accept|Reject|Challenge)|rlm_|eap_(peap|tls|ttls|md5)|\(\d+\) `)
)

// Detect guesses the log format. NPS is checked first: NPS event text never
// contains FreeRADIUS markers, while pasted mixtures should lean NPS.
func Detect(text string) Format {
	if npsRe.MatchString(text) {
		return FormatNPS
	}
	if frRe.MatchString(text) {
		return FormatFreeRADIUS
	}
	return FormatUnknown
}

// npsFields maps labels found in 6273/6274 event text to summary keys.
var npsFields = map[string]string{
	"Reason Code":                    "reason_code",
	"Reason":                         "reason",
	"Authentication Type":            "authentication_type",
	"EAP Type":                       "eap_type",
	"Network Policy Name":            "network_policy",
	"Connection Request Policy Name": "connection_request_policy",
	"NAS Port-Type":                  "nas_port_type",
	"Event ID":                       "event_id",
}

var errorLineRe = regexp.MustCompile(`(?i)\bERROR\b`)
var sentReplyRe = regexp.MustCompile(`Sent (Access-\w+)`)

// Summary extracts a few headline fields used for display; it never includes
// credentials. Both implementations must produce the same keys.
func Summary(text string, f Format) map[string]string {
	out := map[string]string{}
	lines := strings.Split(text, "\n")
	switch f {
	case FormatNPS:
		for _, line := range lines {
			label, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			label = strings.TrimSpace(label)
			value = strings.TrimSpace(value)
			if key, known := npsFields[label]; known && value != "" && value != "-" {
				if _, dup := out[key]; !dup {
					out[key] = value
				}
			}
		}
	case FormatFreeRADIUS:
		errors := 0
		firstError := ""
		finalReply := ""
		for _, line := range lines {
			if errorLineRe.MatchString(line) {
				errors++
				if firstError == "" {
					firstError = strings.TrimSpace(line)
				}
			}
			if m := sentReplyRe.FindStringSubmatch(line); m != nil {
				finalReply = m[1]
			}
		}
		if errors > 0 {
			out["error_lines"] = strconv.Itoa(errors)
			out["first_error"] = firstError
		}
		if finalReply != "" {
			out["final_reply"] = finalReply
		}
	}
	return out
}
