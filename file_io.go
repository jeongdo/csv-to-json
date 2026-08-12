package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const previewRows = 8

type FileSummary struct {
	Path          string     `json:"path"`
	Name          string     `json:"name"`
	Size          int64      `json:"size"`
	Delimiter     string     `json:"delimiter"`
	DelimiterRune string     `json:"delimiterRune"`
	Columns       int        `json:"columns"`
	Headers       []string   `json:"headers"`
	Preview       [][]string `json:"preview"`
}

type Progress struct {
	Percent int   `json:"percent"`
	Bytes   int64 `json:"bytes"`
	Total   int64 `json:"total"`
}

func inspectFile(path string) (*FileSummary, error) {
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
	if err != nil {
		return nil, fmt.Errorf("FILE_READ_FAILED: %w", err)
	}
	if info.IsDir() {
		return nil, errors.New("FILE_READ_FAILED: selected path is a directory")
	}

	delim, ambiguous := detectDelimiterAdvanced(file)
	if ambiguous {
		return nil, errors.New("MIXED_DELIMITER_DETECTED")
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("FILE_READ_FAILED: %w", err)
	}

	reader := csv.NewReader(file)
	reader.Comma = delim
	reader.FieldsPerRecord = 0
	rawHeaders, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("HEADER_READ_FAILED: %w", err)
	}
	headers, err := normaliseHeaders(rawHeaders)
	if err != nil {
		return nil, err
	}

	preview := make([][]string, 0, previewRows)
	for len(preview) < previewRows {
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("CSV_PARSE_FAILED: %w", readErr)
		}
		row := make([]string, len(headers))
		copy(row, record)
		preview = append(preview, row)
	}

	return &FileSummary{
		Path:          absolutePath,
		Name:          info.Name(),
		Size:          info.Size(),
		Delimiter:     delimiterName(delim),
		DelimiterRune: string(delim),
		Columns:       len(headers),
		Headers:       headers,
		Preview:       preview,
	}, nil
}

type progressReader struct {
	r       io.Reader
	total   int64
	read    int64
	lastPct int
	emit    func(Progress)
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if n > 0 {
		r.read += int64(n)
		pct := 100
		if r.total > 0 {
			pct = int(r.read * 100 / r.total)
			if pct > 100 {
				pct = 100
			}
		}
		if pct >= r.lastPct+1 || pct == 100 {
			r.lastPct = pct
			if r.emit != nil {
				r.emit(Progress{Percent: pct, Bytes: r.read, Total: r.total})
			}
		}
	}
	return n, err
}

func sameFilePath(a, b string) bool {
	aAbs, aErr := filepath.Abs(filepath.Clean(a))
	bAbs, bErr := filepath.Abs(filepath.Clean(b))
	if aErr == nil && bErr == nil {
		if runtime.GOOS == "windows" {
			if strings.EqualFold(aAbs, bAbs) {
				return true
			}
		} else if aAbs == bAbs {
			return true
		}
	}
	aInfo, aErr := os.Stat(a)
	bInfo, bErr := os.Stat(b)
	return aErr == nil && bErr == nil && os.SameFile(aInfo, bInfo)
}

func convertFilePath(inputPath, outputPath string, settings ConvertSettings, emit func(Progress)) (ConvertStats, int64, error) {
	if sameFilePath(inputPath, outputPath) {
		return ConvertStats{}, 0, errors.New("OUTPUT_EQUALS_INPUT")
	}

	input, err := os.Open(inputPath)
	if err != nil {
		return ConvertStats{}, 0, fmt.Errorf("FILE_READ_FAILED: %w", err)
	}
	defer input.Close()

	inputInfo, err := input.Stat()
	if err != nil {
		return ConvertStats{}, 0, fmt.Errorf("FILE_READ_FAILED: %w", err)
	}
	if inputInfo.IsDir() {
		return ConvertStats{}, 0, errors.New("FILE_READ_FAILED: selected path is a directory")
	}

	delim, ambiguous := detectDelimiterAdvanced(input)
	if ambiguous {
		return ConvertStats{}, 0, errors.New("MIXED_DELIMITER_DETECTED")
	}
	if _, err = input.Seek(0, io.SeekStart); err != nil {
		return ConvertStats{}, 0, fmt.Errorf("FILE_READ_FAILED: %w", err)
	}

	outputPath, err = filepath.Abs(filepath.Clean(outputPath))
	if err != nil {
		return ConvertStats{}, 0, fmt.Errorf("OUTPUT_CREATE_FAILED: %w", err)
	}
	outputDir := filepath.Dir(outputPath)
	if err = os.MkdirAll(outputDir, 0o755); err != nil {
		return ConvertStats{}, 0, fmt.Errorf("OUTPUT_CREATE_FAILED: %w", err)
	}

	tmp, err := os.CreateTemp(outputDir, ".csv-to-json-*.tmp")
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
		emit(Progress{Percent: 0, Total: inputInfo.Size()})
	}
	progress := &progressReader{r: input, total: inputInfo.Size(), emit: emit}
	stats, err := convertCSVWithSettings(progress, tmp, delim, settings)
	if err != nil {
		return stats, 0, errors.New(mapConvertError(err))
	}
	if err = tmp.Sync(); err != nil {
		return stats, 0, fmt.Errorf("OUTPUT_WRITE_FAILED: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return stats, 0, fmt.Errorf("OUTPUT_WRITE_FAILED: %w", err)
	}

	outputInfo, err := os.Stat(tmpPath)
	if err != nil {
		return stats, 0, fmt.Errorf("OUTPUT_WRITE_FAILED: %w", err)
	}
	if err = replaceOutputFile(tmpPath, outputPath); err != nil {
		return stats, 0, fmt.Errorf("OUTPUT_WRITE_FAILED: %w", err)
	}
	committed = true
	if emit != nil {
		emit(Progress{Percent: 100, Bytes: inputInfo.Size(), Total: inputInfo.Size()})
	}
	return stats, outputInfo.Size(), nil
}

func replaceOutputFile(tempPath, targetPath string) error {
	if _, err := os.Stat(targetPath); errors.Is(err, os.ErrNotExist) {
		return os.Rename(tempPath, targetPath)
	} else if err != nil {
		return err
	}

	backup, err := os.CreateTemp(filepath.Dir(targetPath), ".csv-to-json-backup-*.tmp")
	if err != nil {
		return err
	}
	backupPath := backup.Name()
	_ = backup.Close()
	_ = os.Remove(backupPath)

	if err = os.Rename(targetPath, backupPath); err != nil {
		return err
	}
	if err = os.Rename(tempPath, targetPath); err != nil {
		_ = os.Rename(backupPath, targetPath)
		return err
	}
	_ = os.Remove(backupPath)
	return nil
}
