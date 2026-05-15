package queries

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/netip"
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	minDashes     = 3
	fallbackWidth = 80

	// ANSI color codes
	colorReset     = "\033[0m"
	colorScalarKey = "\033[36m" // cyan   — key with a scalar value
	colorNestedKey = "\033[33m" // yellow — key with an object/array value
	colorDashes    = "\033[90m" // dark grey
	colorValue     = "\033[97m" // bright white
)

// PrettyPrint nicely dumps first-party query results to the terminal
// to avoid circular dependencies, we take query type as int rather than iota name
// qT 0 = First
// qT 1 = Third
// qT 2 = Full
// Future versions may attempt to decouple flag config types from the cli parsing package,
// or general orchestration out of cli.init<thing> functions to avoid this hack
// and allow intuitive dependency usage
func PrettyPrint(ip netip.Addr, queryType int, humanFriendly bool) {
	var res []byte
	var err error

	switch queryType {
	case 0:
		res, err = FirstPartyQuery(ip)
	case 1:
		res, err = ThirdPartyQuery(ip)
	case 2:
		res, err = FullQuery(ip)
	default:
		log.Fatalf("Unknown query type: %d", queryType)
	}
	if err != nil {
		log.Fatalf("Query type %d: %v\n", queryType, err)
	}

	var pretty string

	if !humanFriendly {
		pretty, err = prettyJSON(res)

	} else {
		pretty, err = formatForHumans(res)
	}
	if err != nil {
		log.Fatalf("prettyJson/formatForHumans: %v\n", err)
	}

	fmt.Println(string(pretty))
}

// prettyJSON indents JSON for readability
func prettyJSON(data []byte) (string, error) {
	var out bytes.Buffer
	if err := json.Indent(&out, data, "", "  "); err != nil {
		return "", err
	}
	return string(out.Bytes()), nil
}

func formatForHumans(data []byte) (string, error) {
	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	width := termWidth()

	var sb strings.Builder
	formatValue(&sb, parsed, 0, width)
	return sb.String(), nil
}

// termWidth returns the current terminal column count, or
// fallbackWidth if value is larger than fbW or undetected
func termWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 && w < fallbackWidth {
		return w
	}
	return fallbackWidth
}

// formatValue builds printable strings from arbitrary json, recursing on nested values.
func formatValue(sb *strings.Builder, v any, depth, width int) {
	indent := strings.Repeat("    ", depth)

	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			switch cv := child.(type) {
			case map[string]any:
				sb.WriteString(fmt.Sprintf("%s%s%s%s:\n", indent, colorNestedKey, k, colorReset))
				formatValue(sb, cv, depth+1, width)

			case []any:
				sb.WriteString(fmt.Sprintf("%s%s%s%s:\n", indent, colorNestedKey, k, colorReset))
				for i, item := range cv {
					switch iv := item.(type) {
					case map[string]any:
						// Array index acting as a nested header
						sb.WriteString(fmt.Sprintf("%s    %s[%d]%s:\n", indent, colorNestedKey, i, colorReset))
						formatValue(sb, iv, depth+2, width)
					default:
						scalar := fmt.Sprintf("%v", iv)
						label := fmt.Sprintf("[%d]", i)
						dashes := dashFill(indent+"    ", label, scalar, width)
						sb.WriteString(fmt.Sprintf(
							"%s    %s%s%s %s%s%s %s%s%s\n",
							indent,
							colorScalarKey, label, colorReset,
							colorDashes, dashes, colorReset,
							colorValue, scalar, colorReset,
						))
					}
				}

			default:
				scalar := fmt.Sprintf("%v", cv)
				dashes := dashFill(indent, k, scalar, width)
				sb.WriteString(fmt.Sprintf(
					"%s%s%s%s %s%s%s %s%s%s\n",
					indent,
					colorScalarKey, k, colorReset,
					colorDashes, dashes, colorReset,
					colorValue, scalar, colorReset,
				))
			}
		}

	default:
		sb.WriteString(fmt.Sprintf("%s%s%v%s\n", indent, colorValue, val, colorReset))
	}
}

// dashFill returns a dash string that fills the space between key and value
// so the whole line is exactly `width` columns wide.
//
//	"{indent}{key} {dashes} {value}"
func dashFill(indent, key, value string, width int) string {
	// 2 spaces: one between key and dashes, one between dashes and value
	used := len(indent) + len(key) + 2 + len(value)
	n := width - used
	if n < minDashes {
		n = minDashes
	}
	return strings.Repeat("-", n)
}
