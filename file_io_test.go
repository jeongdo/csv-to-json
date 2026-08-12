package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectFileUsesBoundedPreview(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.csv")
	var b strings.Builder
	b.WriteString("name,age\n")
	for i := 0; i < 20; i++ {
		b.WriteString("Kim,42\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		t.Fatal(err)
	}

	info, err := inspectFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Columns != 2 || info.Delimiter != "comma" || len(info.Preview) != previewRows {
		t.Fatalf("unexpected summary: %+v", info)
	}
	if !filepath.IsAbs(info.Path) {
		t.Fatalf("expected absolute path, got %q", info.Path)
	}
}

func TestConvertFilePathStreamsAndCommitsValidOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.csv")
	output := filepath.Join(dir, "output.json")
	const rows = 5000
	var b strings.Builder
	b.WriteString("id,code,active\n")
	for i := 0; i < rows; i++ {
		b.WriteString("123456789012345678901234567890,01001,true\n")
	}
	if err := os.WriteFile(input, []byte(b.String()), 0o600); err != nil {
		t.Fatal(err)
	}

	progressEvents := 0
	stats, bytesWritten, err := convertFilePath(input, output, defaultConvertSettings(), func(p Progress) {
		progressEvents++
		if p.Percent < 0 || p.Percent > 100 {
			t.Fatalf("invalid progress: %+v", p)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Rows != rows || stats.Columns != 3 || bytesWritten <= 0 || progressEvents == 0 {
		t.Fatalf("unexpected result stats=%+v bytes=%d progress=%d", stats, bytesWritten, progressEvents)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatal("output is not valid JSON")
	}
	if !strings.Contains(string(data), `"id": 123456789012345678901234567890`) {
		t.Fatal("large integer precision was not preserved")
	}
}

func TestConvertFilePathFailureLeavesExistingOutputUntouched(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "bad.csv")
	output := filepath.Join(dir, "output.json")
	if err := os.WriteFile(input, []byte("a,b\n1,2\n3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	const original = `{"keep":"original"}`
	if err := os.WriteFile(output, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	_, _, err := convertFilePath(input, output, defaultConvertSettings(), nil)
	if err == nil || !strings.Contains(err.Error(), "CSV_PARSE_FAILED") {
		t.Fatalf("expected parse failure, got %v", err)
	}
	got, readErr := os.ReadFile(output)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != original {
		t.Fatalf("existing output was changed on failure: %q", got)
	}
}

func TestConvertFilePathAtomicallyReplacesExistingOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "good.csv")
	output := filepath.Join(dir, "output.json")
	if err := os.WriteFile(input, []byte("name,age\nKim,42\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, err := convertFilePath(input, output, defaultConvertSettings(), nil); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(got) || strings.Contains(string(got), "old") {
		t.Fatalf("unexpected replacement output: %s", got)
	}
	backups, err := filepath.Glob(filepath.Join(dir, ".csv-to-json-backup-*.tmp"))
	if err != nil || len(backups) != 0 {
		t.Fatalf("backup leak: %v err=%v", backups, err)
	}
}

func TestConvertFilePathRejectsSameInputOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "same.csv")
	if err := os.WriteFile(input, []byte("a\n1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := convertFilePath(input, input, defaultConvertSettings(), nil)
	if err == nil || err.Error() != "OUTPUT_EQUALS_INPUT" {
		t.Fatalf("expected OUTPUT_EQUALS_INPUT, got %v", err)
	}
}
