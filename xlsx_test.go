package main

import "testing"

func TestXLSXColumnIndex(t *testing.T) {
	cases := map[string]int{"A1": 0, "Z9": 25, "AA2": 26, "AB100": 27}
	for ref, want := range cases {
		if got := xlsxColumnIndex(ref); got != want {
			t.Fatalf("%s got %d want %d", ref, got, want)
		}
	}
}

func TestXLSXCellText(t *testing.T) {
	shared := []string{"hello"}
	got, err := xlsxCellText(xlsxCell{Type: "s", Value: "0"}, shared)
	if err != nil || got != "hello" {
		t.Fatalf("got=%q err=%v", got, err)
	}
	got, err = xlsxCellText(xlsxCell{Type: "b", Value: "1"}, shared)
	if err != nil || got != "true" {
		t.Fatalf("got=%q err=%v", got, err)
	}
}
