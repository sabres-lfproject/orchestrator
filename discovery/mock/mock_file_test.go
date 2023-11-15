package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	inventory "pulwar.isi.edu/sabres/orchestrator/inventory/protocol"
)

func TestLoadingServiceConfig(t *testing.T) {

	_, filename, _, _ := runtime.Caller(0)
	base := filepath.Dir(filename)
	files, err := ioutil.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}

	for _, fi := range files {
		test_file := ""
		if strings.Contains(fi.Name(), "json") {
			// vim swp file
			if strings.Contains(fi.Name(), "swp") {
				continue
			}
			test_file = fi.Name()
		} else {
			continue
		}

		configPath := fmt.Sprintf("%s/%s", base, test_file)

		t.Logf("%s\n", configPath)

		data, err := os.Open(configPath)
		if err != nil {
			t.Fatal(err)
		}
		jsonData, err := io.ReadAll(data)
		if err != nil {
			t.Fatal(err)
		}

		if strings.Contains(fi.Name(), "resource") {
			rec := make([]inventory.ResourceItem, 0)
			err = json.Unmarshal(jsonData, &rec)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("%#v\n", rec)
		} else if strings.Contains(fi.Name(), "slice") {
			// TODO: Load in https://github.com/onosproject/aether-roc-api/blob/master/api/aether-2.1.0-openapi3.yaml
			// https://github.com/nytimes/openapi2proto

			//t.Logf("%#v\n", string(jsonData))
		}
	}
}
