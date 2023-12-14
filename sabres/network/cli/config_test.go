package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"pulwar.isi.edu/sabres/orchestrator/inventory/pkg"
)

func TestLoadingConfig(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	base := filepath.Dir(filename)
	files, err := ioutil.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}

	for _, fi := range files {
		test_file := ""
		if strings.Contains(fi.Name(), "test_config") && !strings.Contains(fi.Name(), ".swp") {
			test_file = fi.Name()
		} else {
			continue
		}

		configPath := fmt.Sprintf("%s/%s", base, test_file)

		t.Logf("%s\n", configPath)

		cfg, err := pkg.LoadInventoryItemConfig(configPath)
		if err != nil {
			if strings.Contains(fi.Name(), "bad") {
				t.Logf("bad config verified\n")
			} else {
				t.Fatalf("%v", err)
			}
		} else {
			for _, config := range cfg {
				t.Logf("%v\n", config)
			}
		}
	}
}
