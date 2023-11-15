package pkg

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
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
		if strings.Contains(fi.Name(), "test_service_config") {
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

		cfg, err := LoadServicesConfig(configPath)
		if err != nil {
			if strings.Contains(fi.Name(), "bad") {
				t.Logf("Found failures\n")
			} else {
				t.Fatalf("%v", err)
			}
		} else {
			for _, ep := range cfg {
				t.Logf("%#v\n", ep)
			}
		}
	}
}
