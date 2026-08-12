package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const delimiterProbeRecords = 32

type ErrorResponse struct {
	Code   string `json:"code"`
	Detail string `json:"detail,omitempty"`
}

type ConvertStats struct {
	Rows      int   `json:"rows"`
	Columns   int   `json:"columns"`
	Delimiter rune  `json:"-"`
	Bytes     int64 `json:"bytes,omitempty"`
}

func writeError(w http.ResponseWriter, status int, code, detail string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Code: code, Detail: detail})
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

func inferValue(s string) any {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		// Preserve explicitly quoted/padded string data instead of silently trimming it.
		if s != "" {
			return s
		}
		return ""
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
	if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
		if !strings.ContainsAny(trimmed, "Nn") &&
			!strings.Contains(trimmed, "Inf") &&
			!strings.Contains(trimmed, "inf") {
			return f
		}
	}
	return s
}

// inferValueJSON remains as a narrow compatibility helper for existing callers/tests.
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

func convertCSV(r io.Reader, w io.Writer, delim rune) (ConvertStats, error) {
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
t		var raw string
			if i < len(record) {
				raw = record[i]
			}
			if err = writeJSONValue(w, inferValue(raw)); err != nil {
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
			// A malformed later row must not erase a delimiter that was
			// already established by a valid header/prefix. The converter
			// will report the parse error during the full pass.
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
	defer rs.Seek(0, io.SeekStart) // best-effort rewind for caller

	candidates := []rune{',', '\t', '|', ';'}
	valid := make([]delimiterScore, 0, len(candidates))
	for _, delim := range candidates {
		score, ok := probeDelimiter(rs, delim)
		if ok && score.columns > 1 {
			valid = append(valid, score)
		}
	}

	if len(valid) == 0 {
		// A delimiter-free file is a valid one-column CSV.
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

func isValidDelimiter(lines []string, delim rune) bool {
	r := strings.NewReader(strings.Join(lines, "\n"))
	score, ok := probeDelimiter(r, delim)
	return ok && score.columns > 0
}

func mapConvertError(err error) (code, detail string) {
	if err == nil {
		return "", ""
	}
	msg := err.Error()
	switch {
	case msg == "EMPTY_HEADER":
		return "EMPTY_HEADER", ""
	case strings.HasPrefix(msg, "DUPLICATE_HEADER:"):
		return "DUPLICATE_HEADER", strings.TrimPrefix(msg, "DUPLICATE_HEADER:")
	case strings.HasPrefix(msg, "HEADER_READ_FAILED"):
		return "HEADER_READ_FAILED", ""
	case strings.HasPrefix(msg, "CSV_PARSE_FAILED"):
		return "CSV_PARSE_FAILED", ""
	default:
		return "UNKNOWN_ERROR", ""
	}
}

func resetMultipartFile(file multipart.File) error {
	_, err := file.Seek(0, io.SeekStart)
	return err
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "INVALID_METHOD", "")
		return
	}
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "UPLOAD_FAILED", "")
		return
	}

	file, _, err := r.FormFile("csvFile")
	if err != nil {
		writeError(w, http.StatusBadRequest, "FILE_READ_FAILED", "")
		return
	}
	defer file.Close()

	delim, ambiguous := detectDelimiterAdvanced(file)
	if ambiguous {
		writeError(w, http.StatusBadRequest, "MIXED_DELIMITER_DETECTED", "")
		return
	}
	if err := resetMultipartFile(file); err != nil {
		writeError(w, http.StatusBadRequest, "FILE_READ_FAILED", "")
		return
	}

	// Convert to a temporary file first. The HTTP response is not committed until
	// the complete CSV has been validated, so malformed late rows cannot produce
	// a successful but truncated JSON download.
	tmp, err := os.CreateTemp("", "csv-to-json-*.json")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "UNKNOWN_ERROR", "")
		return
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		_ = os.Remove(tmpName)
	}()

	if _, err = convertCSV(file, tmp, delim); err != nil {
		code, detail := mapConvertError(err)
		writeError(w, http.StatusBadRequest, code, detail)
		return
	}
	if _, err = tmp.Seek(0, io.SeekStart); err != nil {
		writeError(w, http.StatusInternalServerError, "UNKNOWN_ERROR", "")
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"download.json\"")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = io.Copy(w, tmp)
}
