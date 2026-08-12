package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestJSONToDelimitedFlatArray(t *testing.T) {
	src := `[{"name":"Kim","age":20,"active":true},{"name":"Lee","age":30,"active":false}]`
	headers, rows, err := collectJSONHeaders(strings.NewReader(src))
	if err != nil { t.Fatal(err) }
	if rows != 2 || strings.Join(headers, ",") != "name,age,active" { t.Fatalf("headers=%v rows=%d", headers, rows) }
	var out bytes.Buffer
	stats, err := convertJSONToDelimited(strings.NewReader(src), &out, headers, ',')
	if err != nil { t.Fatal(err) }
	if stats.Rows != 2 { t.Fatalf("rows=%d", stats.Rows) }
	if out.String() != "name,age,active\nKim,20,true\nLee,30,false\n" { t.Fatalf("got %q", out.String()) }
}

func TestJSONToDelimitedUnionHeaders(t *testing.T) {
	src := `[{"a":1},{"b":2,"a":3}]`
	headers, _, err := collectJSONHeaders(strings.NewReader(src))
	if err != nil { t.Fatal(err) }
	if strings.Join(headers, ",") != "a,b" { t.Fatalf("headers=%v", headers) }
	var out bytes.Buffer
	_, err = convertJSONToDelimited(strings.NewReader(src), &out, headers, ',')
	if err != nil { t.Fatal(err) }
	if out.String() != "a,b\n1,\n3,2\n" { t.Fatalf("got %q", out.String()) }
}

func TestJSONToDelimitedSingleObject(t *testing.T) {
	src := `{"id":"01001","note":" a "}`
	headers, rows, err := collectJSONHeaders(strings.NewReader(src))
	if err != nil { t.Fatal(err) }
	if rows != 1 { t.Fatalf("rows=%d", rows) }
	var out bytes.Buffer
	_, err = convertJSONToDelimited(strings.NewReader(src), &out, headers, ';')
	if err != nil { t.Fatal(err) }
	if out.String() != "id;note\n01001;\" a \"\n" { t.Fatalf("got %q", out.String()) }
}

func TestJSONToDelimitedRejectsNested(t *testing.T) {
	src := `[{"user":{"name":"Kim"}}]`
	if _, _, err := collectJSONHeaders(strings.NewReader(src)); err == nil || !strings.Contains(err.Error(), "NESTED_JSON_NOT_SUPPORTED") { t.Fatalf("err=%v", err) }
}

func TestJSONToDelimitedPreservesLargeNumber(t *testing.T) {
	src := `[{"big":92233720368547758081234567890}]`
	headers, _, err := collectJSONHeaders(strings.NewReader(src))
	if err != nil { t.Fatal(err) }
	var out bytes.Buffer
	_, err = convertJSONToDelimited(strings.NewReader(src), &out, headers, ',')
	if err != nil { t.Fatal(err) }
	if !strings.Contains(out.String(), "92233720368547758081234567890") { t.Fatalf("got %q", out.String()) }
}
