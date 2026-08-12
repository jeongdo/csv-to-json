package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestSplitOutputTemplateDefault(t *testing.T) {
	p, s, err := splitOutputTemplate("")
	if err != nil || p != "" || s != "" {
		t.Fatalf("p=%q s=%q err=%v", p, s, err)
	}
}

func TestSplitOutputTemplateWrapper(t *testing.T) {
	p, s, err := splitOutputTemplate(`{"data":{{rows}},"source":"local"}`)
	if err != nil {
		t.Fatal(err)
	}
	if p != `{"data":` || s != `,"source":"local"}` {
		t.Fatalf("p=%q s=%q", p, s)
	}
}

func TestSplitOutputTemplateRejectsMissingOrDuplicateToken(t *testing.T) {
	for _, input := range []string{`{"data":[]}`, `{{rows}}{{rows}}`} {
		if _, _, err := splitOutputTemplate(input); err == nil || !strings.Contains(err.Error(), "OUTPUT_TEMPLATE_TOKEN_REQUIRED") {
			t.Fatalf("input=%q err=%v", input, err)
		}
	}
}

func TestSplitOutputTemplateRejectsTokenInsideString(t *testing.T) {
	if _, _, err := splitOutputTemplate(`{"data":"{{rows}}"}`); err == nil || !strings.Contains(err.Error(), "OUTPUT_TEMPLATE_TOKEN_POSITION") {
		t.Fatalf("err=%v", err)
	}
}

func TestWriteOutputTemplate(t *testing.T) {
	var out bytes.Buffer
	stats, err := writeOutputTemplate(&out, `{"data":{{rows}}}`, func(w io.Writer) (ConvertStats, error) {
		_, err := io.WriteString(w, `[{"id":1}]`)
		return ConvertStats{Rows: 1, Columns: 1}, err
	})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Rows != 1 || out.String() != `{"data":[{"id":1}]}` {
		t.Fatalf("stats=%+v out=%q", stats, out.String())
	}
}

func TestValidateJSONFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "template-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.WriteString(`{"data":[{"id":1},{"id":2}]}`); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	if err = validateJSONFile(f.Name()); err != nil {
		t.Fatal(err)
	}
}

func TestValidateJSONFileRejectsTrailingJSON(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "template-bad-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.WriteString(`[] {}`); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	if err = validateJSONFile(f.Name()); err == nil || !strings.Contains(err.Error(), "OUTPUT_TEMPLATE_INVALID") {
		t.Fatalf("err=%v", err)
	}
}
