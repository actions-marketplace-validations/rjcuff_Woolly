package memory

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	linkA = "20261110-153000-1a2b3c4d"
	linkB = "20261111-090000-deadbeef"
)

func TestParseLinksReadsEveryStyle(t *testing.T) {
	cases := map[string][]string{
		linkA + ", " + linkB:                         {linkA, linkB},
		"[[" + linkA + "]], [[" + linkB + "]]":       {linkA, linkB},
		`["[[` + linkA + `]]", "[[` + linkB + `]]"]`: {linkA, linkB},
		`['[[` + linkA + `]]']`:                      {linkA},
		"[" + linkA + ", " + linkB + "]":             {linkA, linkB},
		"[[ " + linkA + " ]]":                        {linkA},
		`["[[not an id]]", "[[` + linkA + `]]"]`:     {linkA},
		"[[../evil]]":                                nil,
		"":                                           nil,
	}
	for in, want := range cases {
		if got := parseLinks(in); !reflect.DeepEqual(got, want) {
			t.Errorf("parseLinks(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestFormatLinksIsValidYAMLForObsidian(t *testing.T) {
	got := formatLinks([]string{linkA, linkB}, true)
	want := `["[[` + linkA + `]]", "[[` + linkB + `]]"]`
	if got != want {
		t.Fatalf("formatLinks = %s, want %s", got, want)
	}
	if got := formatLinks([]string{linkA}, false); got != linkA {
		t.Fatalf("plain = %s", got)
	}
	if got := formatLinks(nil, true); got != "" {
		t.Fatalf("no links = %q", got)
	}
}

func TestLinksRoundTripInEachStyle(t *testing.T) {
	now := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
	m := &Memory{ID: "20261122-000000-00000001", Content: "x", Source: SourceUser, Status: StatusActive, CreatedAt: now, UpdatedAt: now, Links: []string{linkA, linkB}}
	for _, wiki := range []bool{false, true} {
		got, err := decode(encodeAs(m, wiki))
		if err != nil {
			t.Fatalf("wiki=%v: %v", wiki, err)
		}
		if !reflect.DeepEqual(got.Links, m.Links) {
			t.Errorf("wiki=%v: links = %v, want %v", wiki, got.Links, m.Links)
		}
	}
}

func TestStorePicksLinkStyle(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, "memories")
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(LinksEnv, "")
	if s.wikilinks() {
		t.Fatal("plain by default outside a vault")
	}
	if err := os.Mkdir(filepath.Join(home, ".obsidian"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !s.wikilinks() {
		t.Fatal("WOOLLY_HOME opened as a vault should use wikilinks")
	}
	t.Setenv(LinksEnv, "plain")
	if s.wikilinks() {
		t.Fatal("WOOLLY_LINKS=plain should win over the vault")
	}
	t.Setenv(LinksEnv, "WikiLinks")
	os.Remove(filepath.Join(home, ".obsidian"))
	if !s.wikilinks() {
		t.Fatal("WOOLLY_LINKS=wikilinks should turn it on anywhere")
	}
	m := &Memory{ID: "20261122-000000-00000001", Content: "x", Links: []string{linkA}}
	if !strings.Contains(string(s.encode(m)), `links: ["[[`+linkA+`]]"]`) {
		t.Fatalf("store encode:\n%s", s.encode(m))
	}
}
