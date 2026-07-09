// radius-analyze diagnoses FreeRADIUS debug output (radiusd -X) and Windows
// NPS event log text. Analysis is fully local; nothing is sent anywhere.
//
// Usage:
//
//	radius-analyze [flags] [file]
//	radiusd -X 2>&1 | radius-analyze
//
// With no file argument (or "-"), input is read from stdin.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	analyzer "github.com/authhound/radius-log-analyser"
	"github.com/authhound/radius-log-analyser/internal/diagnose"
	"github.com/authhound/radius-log-analyser/internal/parse"
)

func main() {
	jsonOut := flag.Bool("json", false, "emit the diagnosis as JSON")
	formatFlag := flag.String("format", "auto", "input format: auto, freeradius, or nps")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: radius-analyze [flags] [file]\n\nDiagnoses FreeRADIUS debug output (radiusd -X) and Windows NPS event log text.\nReads the given file, or stdin when no file (or \"-\") is given. Analysis is\nfully local; nothing leaves this machine.\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	text, err := readInput(flag.Arg(0))
	if err != nil {
		fatal(err)
	}
	if len(text) == 0 {
		fatal(fmt.Errorf("no input: pass a file or pipe log text on stdin"))
	}

	var format parse.Format
	switch *formatFlag {
	case "auto":
		format = diagnose.FormatAuto
	case "freeradius", "nps":
		format = parse.Format(*formatFlag)
	default:
		fatal(fmt.Errorf("invalid -format %q (want auto, freeradius, or nps)", *formatFlag))
	}

	eng, err := diagnose.New(analyzer.RulesFS)
	if err != nil {
		fatal(err)
	}
	d := eng.Analyze(string(text), format)

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(d); err != nil {
			fatal(err)
		}
		return
	}
	printText(d)
	if d.Format == string(parse.FormatUnknown) {
		os.Exit(2)
	}
}

func readInput(arg string) ([]byte, error) {
	if arg == "" || arg == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(arg)
}

func printText(d diagnose.Diagnosis) {
	switch d.Format {
	case "freeradius":
		fmt.Println("Format detected : FreeRADIUS debug output (radiusd -X)")
	case "nps":
		fmt.Println("Format detected : Windows NPS event log")
	default:
		fmt.Println("Format detected : unknown")
		fmt.Println()
		fmt.Println("This doesn't look like FreeRADIUS debug output or NPS event text.")
		fmt.Println("For FreeRADIUS, run `radiusd -X` and paste its output; for NPS, use")
		fmt.Println("Event Viewer's 'Copy details as text' on an event 6273/6274.")
		return
	}
	for _, k := range []string{"reason_code", "authentication_type", "eap_type", "network_policy", "final_reply", "error_lines"} {
		if v, ok := d.Summary[k]; ok {
			fmt.Printf("%-16s: %s\n", k, v)
		}
	}
	fmt.Println()

	if len(d.Matches) == 0 {
		fmt.Println("No known failure signature matched.")
		fmt.Println("If this log does show a failure, please open an issue with a redacted")
		fmt.Println("sample so a rule can be added.")
		return
	}

	top := d.Matches[0]
	fmt.Printf("DIAGNOSIS — %s\n", top.Title)
	fmt.Printf("Stage: %s\n\n", top.FailureStage)
	fmt.Println(wrap(top.PlainEnglish, 78))
	fmt.Println()
	fmt.Println("Likely causes, most likely first:")
	for i, c := range top.Causes {
		fmt.Printf("  %d. %s\n", i+1, wrapIndent(c.Cause, 78, "     "))
		fmt.Printf("     -> Next check: %s\n", wrapIndent(c.NextCheck, 78, "        "))
	}
	if len(top.Evidence) > 0 {
		fmt.Println()
		fmt.Println("Evidence:")
		for _, ev := range top.Evidence {
			fmt.Printf("  | %s\n", ev)
		}
	}
	if len(d.Matches) > 1 {
		fmt.Println()
		fmt.Println("Also matched:")
		for _, m := range d.Matches[1:] {
			fmt.Printf("  - %s (%s)\n", m.Title, m.RuleID)
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "radius-analyze:", err)
	os.Exit(1)
}

// wrap does simple word wrapping to width columns.
func wrap(s string, width int) string {
	return wrapIndent(s, width, "")
}

func wrapIndent(s string, width int, indent string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return s
	}
	var out, line string
	for _, w := range words {
		if line == "" {
			line = w
		} else if len(line)+1+len(w) <= width-len(indent) {
			line += " " + w
		} else {
			out += line + "\n" + indent
			line = w
		}
	}
	return out + line
}

func splitWords(s string) []string {
	var words []string
	cur := ""
	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\t' {
			if cur != "" {
				words = append(words, cur)
				cur = ""
			}
		} else {
			cur += string(r)
		}
	}
	if cur != "" {
		words = append(words, cur)
	}
	return words
}
