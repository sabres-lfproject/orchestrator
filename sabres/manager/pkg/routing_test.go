package pkg

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGraphToPath(t *testing.T) {

	contents, err := ioutil.ReadFile("./demo-test-slice.json")
	if err != nil {
		t.Fatalf("%v", err)
	}

	obj := &Slice{}
	json.Unmarshal(contents, obj)
	log.Infof("%+v\n", obj)

	path, err := CreatePath(obj.Edges)
	if err != nil {
		t.Fatalf("%v", err)
	}

	knownPath1 := []string{"00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000007", "00000000-0000-0000-0000-000000000006", "00000000-0000-0000-0000-000000000005", "00000000-0000-0000-0000-000000000004", "00000000-0000-0000-0000-000000000003", "00000000-0000-0000-0000-000000000001"}
	knownPath2 := []string{"00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000003", "00000000-0000-0000-0000-000000000004", "00000000-0000-0000-0000-000000000005", "00000000-0000-0000-0000-000000000006", "00000000-0000-0000-0000-000000000007", "00000000-0000-0000-0000-000000000002"}

	assert.ElementsMatch(t, knownPath1, path, "paths should contain the same values")

	if path[0] == knownPath1[0] {
		for i, _ := range path {
			if path[i] != knownPath1[i] {
				t.Fatalf("%v\n%v\n\nNOT EQUAL", path, knownPath1)
			}
		}
	} else {
		for i, _ := range path {
			if path[i] != knownPath2[i] {
				t.Fatalf("%v\n%v\n\nNOT EQUAL", path, knownPath2)
			}
		}
	}

	log.Infof("path: %v\n", path)
}
