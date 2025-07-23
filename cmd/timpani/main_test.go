package main

import (
	"path/filepath"
	"testing"
)

func TestFlags(t *testing.T) {
	if len(flags()) == 0 {
		t.Errorf("flags() should never be nil or empty")
	}
}

func TestConfigDirAndFile(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", d)

	got := configFile()
	want := filepath.Join(d, ConfigDirName, ConfigFileName)
	if got.SourceURI() != want {
		t.Errorf("configFile() = %q, want %q", got.SourceURI(), want)
	}
}
