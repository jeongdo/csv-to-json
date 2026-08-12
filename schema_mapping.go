package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ColumnMapping struct {
	Source  string `json:"source"`
	Target  string `json:"target"`
	Include bool   `json:"include"`
}

type resolvedMapping struct {
	Source string
	Target string
	Index  int
}

func resolveMappings(headers []string, mappings []ColumnMapping) ([]resolvedMapping, error) {
	index := make(map[string]int, len(headers))
	for i, h := range headers {
		index[h] = i
	}
	if len(mappings) == 0 {
		out := make([]resolvedMapping, len(headers))
		for i, h := range headers {
			out[i] = resolvedMapping{Source: h, Target: h, Index: i}
		}
		return out, nil
	}
	out := make([]resolvedMapping, 0, len(mappings))
	seenTargets := map[string]struct{}{}
	for _, m := range mappings {
		if !m.Include {
			continue
		}
		i, ok := index[m.Source]
		if !ok {
			return nil, fmt.Errorf("SCHEMA_SOURCE_NOT_FOUND:%s", m.Source)
		}
		target := m.Target
		if target == "" {
			target = m.Source
		}
		if _, ok := seenTargets[target]; ok {
			return nil, fmt.Errorf("DUPLICATE_HEADER:%s,0", target)
		}
		seenTargets[target] = struct{}{}
		out = append(out, resolvedMapping{Source: m.Source, Target: target, Index: i})
	}
	if len(out) == 0 {
		return nil, errors.New("SCHEMA_EMPTY")
	}
	return out, nil
}

func convertCSVWithMappings(r io.Reader, w io.Writer, delim rune, settings ConvertSettings, mappings []ColumnMapping) (ConvertStats, error) {
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
	resolved, err := resolveMappings(headers, mappings)
	if err != nil {
		return ConvertStats{}, err
	}
	stats := ConvertStats{Columns: len(resolved), Delimiter: delim}
	if _, err = io.WriteString(w, "[\n"); err != nil {
		return stats, err
	}
	first := true
	for {
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return stats, fmt.Errorf("CSV_PARSE_FAILED: %w", readErr)
		}
		if !first {
			if _, err = io.WriteString(w, ",\n"); err != nil {
				return stats, err
			}
		}
		first = false
		if _, err = io.WriteString(w, "  {\n"); err != nil {
			return stats, err
		}
		for i, m := range resolved {
			if i > 0 {
				if _, err = io.WriteString(w, ",\n"); err != nil {
					return stats, err
				}
			}
			key, _ := json.Marshal(m.Target)
			if _, err = fmt.Fprintf(w, "    %s: ", key); err != nil {
				return stats, err
			}
			var raw string
			if m.Index < len(record) {
				raw = record[m.Index]
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

func convertCSVFilePathMapped(inputPath, outputPath string, settings ConvertSettings, mappings []ColumnMapping, emit func(Progress)) (ConvertStats, int64, error) {
	if sameFilePath(inputPath, outputPath) {
		return ConvertStats{}, 0, errors.New("OUTPUT_EQUALS_INPUT")
	}
	input, err := os.Open(inputPath)
	if err != nil {
		return ConvertStats{}, 0, fmt.Errorf("FILE_READ_FAILED: %w", err)
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return ConvertStats{}, 0, fmt.Errorf("FILE_READ_FAILED: %w", err)
	}
	delim, amb := detectDelimiterAdvanced(input)
	if amb {
		return ConvertStats{}, 0, errors.New("MIXED_DELIMITER_DETECTED")
	}
	if _, err = input.Seek(0, io.SeekStart); err != nil {
		return ConvertStats{}, 0, err
	}
	return writeTempConverted(inputPath, outputPath, info.Size(), emit, func(w io.Writer) (ConvertStats, error) {
		progress := &progressReader{r: input, total: info.Size(), emit: emit}
		return convertCSVWithMappings(progress, w, delim, settings, mappings)
	})
}

func convertJSONToDelimitedMapped(r io.Reader, w io.Writer, headers []string, mappings []ColumnMapping, delim rune) (ConvertStats, error) {
	resolved, err := resolveMappings(headers, mappings)
	if err != nil {
		return ConvertStats{}, err
	}
	writer := csv.NewWriter(w)
	writer.Comma = delim
	outHeaders := make([]string, len(resolved))
	for i, m := range resolved {
		outHeaders[i] = m.Target
	}
	if err = writer.Write(outHeaders); err != nil {
		return ConvertStats{}, err
	}
	stats := ConvertStats{Columns: len(resolved), Delimiter: delim}
	err = walkJSONObjects(r, func(obj flatJSONObject) error {
		row := make([]string, len(resolved))
		for i, m := range resolved {
			row[i] = obj.Values[m.Source]
		}
		if err := writer.Write(row); err != nil {
			return err
		}
		stats.Rows++
		return nil
	})
	writer.Flush()
	if err != nil {
		return stats, err
	}
	return stats, writer.Error()
}

func convertJSONFilePathMapped(inputPath, outputPath, delimiterName string, mappings []ColumnMapping, emit func(Progress)) (ConvertStats, int64, error) {
	if sameFilePath(inputPath, outputPath) {
		return ConvertStats{}, 0, errors.New("OUTPUT_EQUALS_INPUT")
	}
	input, err := os.Open(inputPath)
	if err != nil {
		return ConvertStats{}, 0, fmt.Errorf("FILE_READ_FAILED: %w", err)
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return ConvertStats{}, 0, err
	}
	headers, _, err := collectJSONHeaders(input)
	if err != nil {
		return ConvertStats{}, 0, err
	}
	if _, err = input.Seek(0, io.SeekStart); err != nil {
		return ConvertStats{}, 0, err
	}
	return writeTempConverted(inputPath, outputPath, info.Size(), emit, func(w io.Writer) (ConvertStats, error) {
		progress := &progressReader{r: input, total: info.Size(), emit: emit}
		return convertJSONToDelimitedMapped(progress, w, headers, mappings, csvDelimiterFromName(delimiterName))
	})
}

func convertXLSXToJSONMapped(filename string, w io.Writer, settings ConvertSettings, mappings []ColumnMapping) (ConvertStats, error) {
	var headers []string
	var resolved []resolvedMapping
	stats := ConvertStats{}
	first := true
	err := walkXLSXRows(filename, func(row []string) error {
		if headers == nil {
			var err error
			headers, err = normaliseHeaders(row)
			if err != nil {
				return err
			}
			resolved, err = resolveMappings(headers, mappings)
			if err != nil {
				return err
			}
			stats.Columns = len(resolved)
			_, err = io.WriteString(w, "[\n")
			return err
		}
		if !first {
			if _, err := io.WriteString(w, ",\n"); err != nil {
				return err
			}
		}
		first = false
		if _, err := io.WriteString(w, "  {\n"); err != nil {
			return err
		}
		for i, m := range resolved {
			if i > 0 {
				if _, err := io.WriteString(w, ",\n"); err != nil {
					return err
				}
			}
			key, _ := json.Marshal(m.Target)
			if _, err := fmt.Fprintf(w, "    %s: ", key); err != nil {
				return err
			}
			var raw string
			if m.Index < len(row) {
				raw = row[m.Index]
			}
			if err := writeJSONValue(w, inferValueWithSettings(raw, settings)); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, "\n  }"); err != nil {
			return err
		}
		stats.Rows++
		return nil
	})
	if err != nil {
		return stats, err
	}
	if headers == nil {
		return stats, errors.New("HEADER_READ_FAILED")
	}
	if stats.Rows == 0 {
		_, err = io.WriteString(w, "]\n")
	} else {
		_, err = io.WriteString(w, "\n]\n")
	}
	return stats, err
}

func convertXLSXFilePathMapped(inputPath, outputPath string, settings ConvertSettings, mappings []ColumnMapping, emit func(Progress)) (ConvertStats, int64, error) {
	info, err := os.Stat(inputPath)
	if err != nil {
		return ConvertStats{}, 0, err
	}
	return writeTempConverted(inputPath, outputPath, info.Size(), emit, func(w io.Writer) (ConvertStats, error) {
		return convertXLSXToJSONMapped(inputPath, w, settings, mappings)
	})
}

func writeTempConverted(inputPath, outputPath string, total int64, emit func(Progress), convert func(io.Writer) (ConvertStats, error)) (ConvertStats, int64, error) {
	if sameFilePath(inputPath, outputPath) {
		return ConvertStats{}, 0, errors.New("OUTPUT_EQUALS_INPUT")
	}
	abs, err := filepath.Abs(filepath.Clean(outputPath))
	if err != nil {
		return ConvertStats{}, 0, fmt.Errorf("OUTPUT_CREATE_FAILED: %w", err)
	}
	if err = os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return ConvertStats{}, 0, fmt.Errorf("OUTPUT_CREATE_FAILED: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(abs), ".data-convert-*.tmp")
	if err != nil {
		return ConvertStats{}, 0, fmt.Errorf("OUTPUT_CREATE_FAILED: %w", err)
	}
	tmpPath := tmp.Name()
	committed := false
	defer func() {
		_ = tmp.Close()
		if !committed {
			_ = os.Remove(tmpPath)
		}
	}()
	if emit != nil {
		emit(Progress{Percent: 0, Total: total})
	}
	stats, err := convert(tmp)
	if err != nil {
		return stats, 0, err
	}
	if err = tmp.Sync(); err != nil {
		return stats, 0, fmt.Errorf("OUTPUT_WRITE_FAILED: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return stats, 0, fmt.Errorf("OUTPUT_WRITE_FAILED: %w", err)
	}
	info, err := os.Stat(tmpPath)
	if err != nil {
		return stats, 0, fmt.Errorf("OUTPUT_WRITE_FAILED: %w", err)
	}
	if err = replaceOutputFile(tmpPath, abs); err != nil {
		return stats, 0, fmt.Errorf("OUTPUT_WRITE_FAILED: %w", err)
	}
	committed = true
	if emit != nil {
		emit(Progress{Percent: 100, Bytes: total, Total: total})
	}
	return stats, info.Size(), nil
}
