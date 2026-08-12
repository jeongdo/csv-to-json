package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInferValueJSONAlwaysValid(t *testing.T) {
	cases := []string{"+1", ".5", "1.", "-01", "01", "0.5", "1e3", "true", "false", "hello", "  hello  ", ""}
	for _, input := range cases {
		got := inferValueJSON(input)
		if !json.Valid([]byte(got)) {
			t.Fatalf("inferValueJSON(%q) produced invalid JSON: %s", input, got)
		}
	}
}

func TestInferValueSemantics(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"+1", "1"},
		{".5", "0.5"},
		{"1.", "1"},
		{"-01", `"-01"`},
		{"01001", `"01001"`},
		{"  hello  ", `"  hello  "`},
		{" 42 ", "42"},
	}
	for _, tt := range tests {
		if got := inferValueJSON(tt.input); got != tt.want {
			t.Errorf("inferValueJSON(%q)=%s, want %s", tt.input, got, tt.want)
		}
	}
}

func TestNormaliseHeaders(t *testing.T) {
	headers, err := normaliseHeaders([]string{"\uFEFF name ", "age"})
	if err != nil || headers[0] != "name" || headers[1] != "age" {
		t.Fatalf("unexpected headers=%v err=%v", headers, err)
	}
	if _, err := normaliseHeaders([]string{"name", " name "}); err == nil || !strings.HasPrefix(err.Error(), "DUPLICATE_HEADER:") {
		t.Fatalf("expected duplicate header error, got %v", err)
	}
	if _, err := normaliseHeaders([]string{"name", "  "}); err == nil || err.Error() != "EMPTY_HEADER" {
		t.Fatalf("expected empty header error, got %v", err)
	}
}

func TestDetectDelimiter(t *testing.T) {
	tests := []struct {
		name      string
		csv       string
		want      rune
		ambiguous bool
	}{
		{"comma", "a,b\n1,2\n", ',', false},
		{"tab", "a\tb\n1\t2\n", '\t', false},
		{"pipe", "a|b\n1|2\n", '|', false},
		{"semicolon", "a;b\n1;2\n", ';', false},
		{"single column", "name\nKim\nLee\n", ',', false},
		{"quoted alternate delimiter", "name,description\nKim,\"a|b;c\"\n", ',', false},
		{"mixed header", "a,b;c\n1,2;3\n", ',', true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ambiguous := detectDelimiterAdvanced(strings.NewReader(tt.csv))
			if got != tt.want || ambiguous != tt.ambiguous {
				t.Fatalf("got delim=%q ambiguous=%v, want %q/%v", got, ambiguous, tt.want, tt.ambiguous)
			}
		})
	}
}

func TestConvertCSVPreservesOrderAndWhitespace(t *testing.T) {
	input := "name,code,note,active\nKim,01001,\"  keep me  \",true\n"
	var out bytes.Buffer
	stats, err := convertCSV(strings.NewReader(input), &out, ',')
	if err != nil {
		t.Fatal(err)
	}
	if stats.Rows != 1 || stats.Columns != 4 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	got := out.String()
	if !strings.Contains(got, `"code": "01001"`) || !strings.Contains(got, `"note": "  keep me  "`) {
		t.Fatalf("data changed unexpectedly:\n%s", got)
	}
	if strings.Index(got, `"name"`) > strings.Index(got, `"code"`) {
		t.Fatalf("column order was not preserved:\n%s", got)
	}
	if !json.Valid(out.Bytes()) {
		t.Fatalf("output is invalid JSON:\n%s", got)
	}
}

func TestConvertHandlerRejectsLateParseErrorWithoutPartialJSON(t *testing.T) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("csvFile", "bad.csv")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(part, "a,b\n1,2\n3\n")
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/convert", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rr := httptest.NewRecorder()
	convertHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response should be JSON error, got %q: %v", rr.Body.String(), err)
	}
	if resp.Code != "CSV_PARSE_FAILED" {
		t.Fatalf("code=%s, want CSV_PARSE_FAILED", resp.Code)
	}
}

func FuzzInferValueJSONValid(f *testing.F) {
	for _, seed := range []string{"+1", ".5", "-01", "hello", "true", "1e309", "  x  "} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if got := inferValueJSON(s); !json.Valid([]byte(got)) {
			t.Fatalf("invalid JSON for %q: %s", s, got)
		}
	})
}
