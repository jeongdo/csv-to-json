package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type InputSummary struct {
	Path          string     `json:"path"`
	Name          string     `json:"name"`
	Size          int64      `json:"size"`
	Kind          string     `json:"kind"`
	Delimiter     string     `json:"delimiter,omitempty"`
	DelimiterRune string     `json:"delimiterRune,omitempty"`
	Columns       int        `json:"columns"`
	Headers       []string   `json:"headers"`
	Preview       [][]string `json:"preview"`
}

type DesktopConvertSettings struct {
	InferTypes      bool   `json:"inferTypes"`
	EmptyAsNull     bool   `json:"emptyAsNull"`
	OutputDelimiter string `json:"outputDelimiter"`
}

func inspectInput(path string) (*InputSummary, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".json" {
		return inspectJSONInput(path)
	}
	summary, err := inspectFile(path)
	if err != nil {
		return nil, err
	}
	return &InputSummary{Path: summary.Path, Name: summary.Name, Size: summary.Size, Kind: "csv", Delimiter: summary.Delimiter, DelimiterRune: summary.DelimiterRune, Columns: summary.Columns, Headers: summary.Headers, Preview: summary.Preview}, nil
}

var errPreviewComplete = errors.New("preview complete")

func inspectJSONInput(path string) (*InputSummary, error) {
	absolutePath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("FILE_READ_FAILED: %w", err)
	}
	file, err := os.Open(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("FILE_READ_FAILED: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		return nil, errors.New("FILE_READ_FAILED")
	}

	headers := make([]string, 0)
	seen := map[string]struct{}{}
	objects := make([]flatJSONObject, 0, previewRows)
	err = walkJSONObjects(file, func(obj flatJSONObject) error {
		for _, key := range obj.Keys {
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				headers = append(headers, key)
			}
		}
		objects = append(objects, obj)
		if len(objects) >= previewRows {
			return errPreviewComplete
		}
		return nil
	})
	if err != nil && !errors.Is(err, errPreviewComplete) {
		return nil, err
	}
	if len(headers) == 0 {
		return nil, errors.New("JSON_EMPTY_NO_COLUMNS")
	}
	preview := make([][]string, len(objects))
	for i, obj := range objects {
		preview[i] = make([]string, len(headers))
		for j, header := range headers {
			preview[i][j] = obj.Values[header]
		}
	}
	return &InputSummary{Path: absolutePath, Name: info.Name(), Size: info.Size(), Kind: "json", Columns: len(headers), Headers: headers, Preview: preview}, nil
}

func convertJSONFilePath(inputPath, outputPath, delimiterName string, emit func(Progress)) (ConvertStats, int64, error) {
	if sameFilePath(inputPath, outputPath) {
		return ConvertStats{}, 0, errors.New("OUTPUT_EQUALS_INPUT")
	}
	input, err := os.Open(inputPath)
	if err != nil {
		return ConvertStats{}, 0, fmt.Errorf("FILE_READ_FAILED: %w", err)
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil || info.IsDir() {
		return ConvertStats{}, 0, errors.New("FILE_READ_FAILED")
	}
	headers, _, err := collectJSONHeaders(input)
	if err != nil {
		return ConvertStats{}, 0, err
	}
	if len(headers) == 0 {
		return ConvertStats{}, 0, errors.New("JSON_EMPTY_NO_COLUMNS")
	}
	if _, err = input.Seek(0, io.SeekStart); err != nil {
		return ConvertStats{}, 0, fmt.Errorf("FILE_READ_FAILED: %w", err)
	}
	outputPath, err = filepath.Abs(filepath.Clean(outputPath))
	if err != nil {
		return ConvertStats{}, 0, fmt.Errorf("OUTPUT_CREATE_FAILED: %w", err)
	}
	if err = os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return ConvertStats{}, 0, fmt.Errorf("OUTPUT_CREATE_FAILED: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(outputPath), ".json-to-csv-*.tmp")
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
		emit(Progress{Percent: 0, Total: info.Size()})
	}
	progress := &progressReader{r: input, total: info.Size(), emit: emit}
	stats, err := convertJSONToDelimited(progress, tmp, headers, csvDelimiterFromName(delimiterName))
	if err != nil {
		return stats, 0, err
	}
	if err = tmp.Sync(); err != nil {
		return stats, 0, fmt.Errorf("OUTPUT_WRITE_FAILED: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return stats, 0, fmt.Errorf("OUTPUT_WRITE_FAILED: %w", err)
	}
	outInfo, err := os.Stat(tmpPath)
	if err != nil {
		return stats, 0, fmt.Errorf("OUTPUT_WRITE_FAILED: %w", err)
	}
	if err = replaceOutputFile(tmpPath, outputPath); err != nil {
		return stats, 0, fmt.Errorf("OUTPUT_WRITE_FAILED: %w", err)
	}
	committed = true
	if emit != nil {
		emit(Progress{Percent: 100, Bytes: info.Size(), Total: info.Size()})
	}
	return stats, outInfo.Size(), nil
}
