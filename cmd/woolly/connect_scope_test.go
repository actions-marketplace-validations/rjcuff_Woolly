package main

import (
	"reflect"
	"testing"
)

func TestScopeFlags(t *testing.T) {
	got, err := scopeFlags("  ann  lee ", "bot")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"--user", "ann lee", "--agent", "bot"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("scopeFlags = %v, want %v", got, want)
	}
	if _, err := scopeFlags("   ", ""); err == nil {
		t.Fatal("a blank --user should be refused")
	}
	if got, err := scopeFlags("", ""); err != nil || got != nil {
		t.Fatalf("no flags = %v, %v; want nil", got, err)
	}
}

func TestConnectRefusesScopeOnDisconnect(t *testing.T) {
	if err := cmdConnect([]string{"--remove", "--agent", "bot", "--dry-run"}); err == nil {
		t.Fatal("disconnect with --agent should be refused")
	}
}
