package pkg

import (
	"testing"
	//"github.com/stretchr/testify/assert"
)

func TestReadingSite(t *testing.T) {
	_, err := ReadXFromFile("site", "site.test.json")
	if err != nil {
		t.Errorf("Failed to read site from test file: %v\n", err)
	}
}

func TestReadingSlice(t *testing.T) {
	_, err := ReadXFromFile("slice", "slice.test.json")
	if err != nil {
		t.Errorf("Failed to read slice from test file: %v\n", err)
	}
}

func TestReadingDeviceGroup(t *testing.T) {
	_, err := ReadXFromFile("device-group", "device-group.test.json")
	if err != nil {
		t.Errorf("Failed to read device group from test file: %v\n", err)
	}
}

func TestReadingDeviceList(t *testing.T) {
	_, err := ReadXFromFile("device", "devices.test.json")
	if err != nil {
		t.Errorf("Failed to read devices from test file: %v\n", err)
	}
}
