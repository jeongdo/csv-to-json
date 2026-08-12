package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const delimiterProbeRecords = 32

type ConvertStats struct {
	Rows      int  `json:"rows"`
	Columns   int  `json:"columns"`
	Delimiter rune `json:"-"`
}

type ConvertSettings struct {
	InferTypes  bool `json:"inferTypes"`
	EmptyAsNull bool `json:"emptyAsNull"`
}

func defaultConvertSettings() ConvertSettings {
	return ConvertSettings{InferTypes: true}
}

func removeBOM(s string) string {
	return strings.TrimPrefix(s, "\uFEFF")
}

func normaliseHeaders(headers []string) ([]string, error) {
	cleaned := make([]string, len(headers))
	seen := make(map[string]struct{}, len(headers))
	duplicates := make([]string, 0)
	duplicateSeen := make(map[string]struct{})

	for i, raw := range headers {
		h := strings.TrimSpace(removeBOM(raw))
		if h == "" {
			return nil, errors.New("EMPTY_HEADER")
		}
		cleaned[i] = h
		if _, ok := seen[h]; ok {
			if _, recorded := duplicateSeen[h]; !recorded {
				duplicates = append(duplicates, h)
				duplicateSeen[h] = struct{}{}
			}
			continue
		}
		seen[h] = struct{}{}
	}

	if len(duplicates) > 0 {
		return nil, fmt.Errorf("DUPLICATE_HEADER:%s,%d", duplicates[0], len(duplicates)-1)
	}
	return cleaned, nil
}

func validateHeaders(headers []string) error {
	_, err := normaliseHeaders(headers)
	return err
}

func hasProtectedLeadingZero(s string) bool {
	if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	return len(s) > 1 && s[0] == '0' && s[1] != '.'
}

func parseJSONNumber(s string) (json.Number, bool) {
	if s == "" || (s[0] != '-' && (s[0] < '0' || s[0] > '9')) {
		return "", false
	}
	if !json.Valid([]byte(s)) {
		return "", false
	}
	return json.Number(s), true
}

func inferValueWithSettings(s string, settings ConvertSettings) any {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		if s == "" && settings.EmptyAsNull {
			return nil
		}
		return s
	}
	if !settings.InferTypes {
		return s
	}
	if trimmed == "true" {
		return true
	}
	if trimmed == "false" {
		return false
	}
	if hasProtectedLeadingZero(trimmed) {
		return s
	}
	if number, ok := parseJSONNumber(trimmed); ok {
		return number
	}
	return s
}

func inferValue(s string) any {
	return inferValueWithSettings(s, defaultConvertSettings())
}

func inferValueJSON(s string) string {
	b, err := json.Marshal(inferValue(s))
	if err != nil {
		return `""`
	}
	return string(b)
}

func writeJSONValue(w io.Writer, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func convertCSVWithSettings(r io.Reader, w io.Writer, delim rune, settings ConvertSettings) (ConvertStats, error) {
	reader := csv.NewReader(r)
	reader.Comma = delim
	reader.FieldsPerRecord = 0

	rawHeaders, err := reader.Read()
	if err != nil {
		return ConvertStats{}, fmt.Errorf("HEADER_READ_FAILED: %w", err)
	}
	headers, err := normaliseHeaders(rawHeaders)
	if err != nil {
		return ConvertStats{}, err
	}

	stats := ConvertStats{Columns: len(headers), Delimiter: delim}
	if _, err = io.WriteString(w, "[\n"); err != nil {
		return stats, err
	}

	firstRecord := true
	for {
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return stats, fmt.Errorf("CSV_PARSE_FAILED: %w", readErr)
		}

		if !firstRecord {
			if _, err = io.WriteString(w, ",\n"); err != nil {
				return stats, err
			}
		}
		firstRecord = false
		if _, err = io.WriteString(w, "  {\n"); err != nil {
			return stats, err
		}

		for i, header := range headers {
			if i > 0 {
				if _, err = io.WriteString(w, ",\n"); err != nil {
					return stats, err
				}
			}
			key, _ := json.Marshal(header)
			if _, err = fmt.Fprintf(w, "    %s: ", key); err != nil {
				return stats, err
			}
			var raw string
			if i < len(record) {
				raw = record[i]
			}
			if err = writeJSONValue(w, inferValueWithSettings(raw, settings)); err != nil {
				return stats, err
			}
		}
		if _, err = io.WriteString(w, "\n  }"); err != nil {
			return stats, err
		}
		stats.Rows++
	}

	if stats.Rows == 0 {
		_, err = io.WriteString(w, "]\n")
	} else {
		_, err = io.WriteString(w, "\n]\n")
	}
	return stats, err
}

func convertCSV(r io.Reader, w io.Writer, delim rune) (ConvertStats, error) {
	return convertCSVWithSettings(r, w, delim, defaultConvertSettings())
}

type delimiterScore struct {
	delim   rune
	columns int
	records int
}

func probeDelimiter(rs io.ReadSeeker, delim rune) (delimiterScore, bool) {
	if _, err := rs.Seek(0, io.SeekStart); err != nil {
		return delimiterScore{}, false
	}
	r := csv.NewReader(rs)
	r.Comma = delim
	r.FieldsPerRecord = 0

	columns := 0
	records := 0
	for records < delimiterProbeRecords {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if records > 0 {
				break
			}
			return delimiterScore{}, false
		}
		if records == 0 {
			columns = len(rec)
		}
		records++
	}
	return delimiterScore{delim: delim, columns: columns, records: records}, records > 0
}

func detectDelimiterAdvanced(r io.Reader) (rune, bool) {
	rs, ok := r.(io.ReadSeeker)
	if !ok {
		return ',', true
	}
	defer func() { _, _ = rs.Seek(0, io.SeekStart) }()

	candidates := []rune{',', '\t', '|', ';'}
	valid := make([]delimiterScore, 0, len(candidates))
	for _, delim := range candidates {
		score, ok := probeDelimiter(rs, delim)
		if ok && score.columns > 1 {
			valid = append(valid, score)
		}
	}

	if len(valid) == 0 {
		if score, ok := probeDelimiter(rs, ','); ok && score.columns == 1 {
			return ',', false
		}
		return ',', true
	}

	best := valid[0]
	ambiguous := false
	for _, candidate := range valid[1:] {
		if candidate.columns > best.columns {
			best = candidate
			ambiguous = false
			continue
		}
		if candidate.columns == best.columns {
			ambiguous = true
		}
	}
	if ambiguous {
		return ',', true
	}
	return best.delim, false
}

func delimiterName(delim rune) string {
	switch delim {
	case ',':
		return "comma"
	case '\t':
		return "tab"
	case '|':
		return "pipe"
	case ';':
		return "semicolon"
	default:
		return string(delim)
	}
}

func mapConvertError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case msg == "EMPTY_HEADER":
		return "EMPTY_HEADER"
	case strings.HasPrefix(msg, "DUPLICATE_HEADER:"):
		return msg
	case strings.HasPrefix(msg, "HEADER_READ_FAILED"):
		return "HEADER_READ_FAILED"
	case strings.HasPrefix(msg, "CSV_PARSE_FAILED"):
		return "CSV_PARSE_FAILED"
	default:
		return "UNKNOWN_ERROR"
	}
}
