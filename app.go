package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx        context.Context
	converting atomic.Bool
}

type ConversionResult struct {
	Cancelled  bool   `json:"cancelled"`
	OutputPath string `json:"outputPath,omitempty"`
	Rows       int    `json:"rows,omitempty"`
	Columns    int    `json:"columns,omitempty"`
	Bytes      int64  `json:"bytes,omitempty"`
	DurationMS int64  `json:"durationMs,omitempty"`
}

func NewApp() *App                         { return &App{} }
func (a *App) startup(ctx context.Context) { a.ctx = ctx }

func (a *App) SelectInput() (*InputSummary, error) {
	path, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Select CSV / JSON / XLSX File",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "Data files (*.csv;*.tsv;*.txt;*.json;*.xlsx)", Pattern: "*.csv;*.tsv;*.txt;*.json;*.xlsx"},
			{DisplayName: "All files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil || path == "" {
		return nil, err
	}
	return inspectInput(path)
}

func (a *App) InspectInput(path string) (*InputSummary, error) { return inspectInput(path) }

func (a *App) ConvertInput(inputPath string, settings DesktopConvertSettings) (*ConversionResult, error) {
	if !a.converting.CompareAndSwap(false, true) {
		return nil, errors.New("CONVERSION_IN_PROGRESS")
	}
	defer a.converting.Store(false)

	summary, err := inspectInput(inputPath)
	if err != nil {
		return nil, err
	}
	inputPath = filepath.Clean(inputPath)
	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, errors.New("FILE_READ_FAILED")
	}

	baseName := strings.TrimSuffix(inputInfo.Name(), filepath.Ext(inputInfo.Name()))
	var outputPath string
	if summary.Kind == "json" {
		outputPath, err = wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
			Title: "Save CSV", DefaultDirectory: filepath.Dir(inputPath), DefaultFilename: baseName + ".csv",
			Filters: []wailsruntime.FileFilter{{DisplayName: "CSV (*.csv)", Pattern: "*.csv"}},
		})
	} else {
		outputPath, err = wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
			Title: "Save JSON", DefaultDirectory: filepath.Dir(inputPath), DefaultFilename: baseName + ".json",
			Filters: []wailsruntime.FileFilter{{DisplayName: "JSON (*.json)", Pattern: "*.json"}},
		})
	}
	if err != nil {
		return nil, err
	}
	if outputPath == "" {
		return &ConversionResult{Cancelled: true}, nil
	}

	started := time.Now()
	var stats ConvertStats
	var outputBytes int64
	if summary.Kind == "json" {
		stats, outputBytes, err = convertJSONFilePathMapped(inputPath, outputPath, settings.OutputDelimiter, settings.Mappings, func(progress Progress) {
			wailsruntime.EventsEmit(a.ctx, "conversion:progress", progress)
		})
	} else if summary.Kind == "xlsx" {
		stats, outputBytes, err = convertXLSXFilePathMapped(inputPath, outputPath, ConvertSettings{InferTypes: settings.InferTypes, EmptyAsNull: settings.EmptyAsNull}, settings.Mappings, func(progress Progress) {
			wailsruntime.EventsEmit(a.ctx, "conversion:progress", progress)
		})
	} else {
		stats, outputBytes, err = convertCSVFilePathMapped(inputPath, outputPath, ConvertSettings{InferTypes: settings.InferTypes, EmptyAsNull: settings.EmptyAsNull}, settings.Mappings, func(progress Progress) {
			wailsruntime.EventsEmit(a.ctx, "conversion:progress", progress)
		})
	}
	if err != nil {
		return nil, err
	}
	outputPath, _ = filepath.Abs(filepath.Clean(outputPath))
	return &ConversionResult{OutputPath: outputPath, Rows: stats.Rows, Columns: stats.Columns, Bytes: outputBytes, DurationMS: time.Since(started).Milliseconds()}, nil
}

// Backward-compatible bindings retained for existing callers.
func (a *App) SelectCSV() (*FileSummary, error) {
	path, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title:   "Select CSV / TSV File",
		Filters: []wailsruntime.FileFilter{{DisplayName: "Delimited text (*.csv;*.tsv;*.txt)", Pattern: "*.csv;*.tsv;*.txt"}},
	})
	if err != nil || path == "" {
		return nil, err
	}
	return inspectFile(path)
}
func (a *App) InspectFile(path string) (*FileSummary, error) { return inspectFile(path) }
func (a *App) ConvertFile(inputPath string, settings ConvertSettings) (*ConversionResult, error) {
	if !a.converting.CompareAndSwap(false, true) {
		return nil, errors.New("CONVERSION_IN_PROGRESS")
	}
	defer a.converting.Store(false)
	inputInfo, err := os.Stat(filepath.Clean(inputPath))
	if err != nil {
		return nil, errors.New("FILE_READ_FAILED")
	}
	base := strings.TrimSuffix(inputInfo.Name(), filepath.Ext(inputInfo.Name())) + ".json"
	outputPath, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{Title: "Save JSON", DefaultDirectory: filepath.Dir(inputPath), DefaultFilename: base, Filters: []wailsruntime.FileFilter{{DisplayName: "JSON (*.json)", Pattern: "*.json"}}})
	if err != nil {
		return nil, err
	}
	if outputPath == "" {
		return &ConversionResult{Cancelled: true}, nil
	}
	started := time.Now()
	stats, bytes, err := convertFilePath(inputPath, outputPath, settings, func(progress Progress) { wailsruntime.EventsEmit(a.ctx, "conversion:progress", progress) })
	if err != nil {
		return nil, err
	}
	outputPath, _ = filepath.Abs(filepath.Clean(outputPath))
	return &ConversionResult{OutputPath: outputPath, Rows: stats.Rows, Columns: stats.Columns, Bytes: bytes, DurationMS: time.Since(started).Milliseconds()}, nil
}

func (a *App) RevealFile(path string) error {
	path = filepath.Clean(path)
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer.exe", "/select,", path).Start()
	case "darwin":
		return exec.Command("open", "-R", path).Start()
	default:
		return exec.Command("xdg-open", filepath.Dir(path)).Start()
	}
}
