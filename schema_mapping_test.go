package main

import (
	"strings"
	"testing"
)

func TestResolveMappingsIdentity(t *testing.T) {
	got, err := resolveMappings([]string{"id", "name"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Target != "id" || got[1].Target != "name" {
		t.Fatalf("got=%+v", got)
	}
}

func TestResolveMappingsRenameDropAndReorder(t *testing.T) {
	got, err := resolveMappings([]string{"id", "name", "age"}, []ColumnMapping{
		{Source: "age", Target: "years", Include: true},
		{Source: "id", Target: "identifier", Include: true},
		{Source: "name", Include: false},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Index != 2 || got[0].Target != "years" || got[1].Index != 0 || got[1].Target != "identifier" {
		t.Fatalf("got=%+v", got)
	}
}

func TestResolveMappingsRejectsDuplicateTargets(t *testing.T) {
	_, err := resolveMappings([]string{"a", "b"}, []ColumnMapping{{Source: "a", Target: "x", Include: true}, {Source: "b", Target: "x", Include: true}})
	if err == nil || !strings.Contains(err.Error(), "DUPLICATE_HEADER") {
		t.Fatalf("err=%v", err)
	}
}

func TestResolveMappingsRejectsEmptySchema(t *testing.T) {
	_, err := resolveMappings([]string{"a"}, []ColumnMapping{{Source: "a", Include: false}})
	if err == nil || err.Error() != "SCHEMA_EMPTY" {
		t.Fatalf("err=%v", err)
	}
}
