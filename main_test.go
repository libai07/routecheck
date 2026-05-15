package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTargetsFromFileRejectsIncompleteTarget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "targets.json")
	if err := os.WriteFile(path, []byte(`[{"name":"北京电信"}]`), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := loadTargetsFromFile(path); err == nil {
		t.Fatal("loadTargetsFromFile incomplete target error = nil")
	}
}

func TestLoadTargetsFromFileSupportsObjectFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "targets.json")
	data := []byte(`{"targets":[{"name":"北京电信","ip":"219.141.140.10"}]}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	targets, err := loadTargetsFromFile(path)
	if err != nil {
		t.Fatalf("loadTargetsFromFile error = %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("len(targets) = %d, want 1", len(targets))
	}
	if targets[0].Name != "北京电信" || targets[0].IP != "219.141.140.10" {
		t.Fatalf("target = %+v", targets[0])
	}
}
