package pkg

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
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
		if strings.Contains(fi.Name(), "test_config") {
			test_file = fi.Name()
		} else {
			continue
		}

		configPath := fmt.Sprintf("%s/%s", base, test_file)

		t.Logf("%s\n", configPath)

		cfg, err := LoadConfig(configPath)
		if err != nil {
			if strings.Contains(fi.Name(), "bad") {
				t.Logf("Found failures\n")
			} else {
				t.Fatalf("%v", err)
			}
		} else {

			_, err = SetEtcdSettings(cfg)
			if err != nil {
				t.Fatalf("%v", err)
			}

			t.Logf("%v\n", cfg.Etcd)
		}
	}
}
