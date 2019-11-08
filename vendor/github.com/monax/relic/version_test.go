package relic

import (
	"testing"
)

func TestVersion(t *testing.T) {
	versionString := "34.5.33"
	version, err := AsVersion(versionString)
	if err != nil {
		t.Error(err)
	}

	// Check parsed version numbers match the version string
	if version.String() != versionString {
		t.Errorf("versionString '%s' should match parsed version '%s'", versionString, version.String())
	}
}

func TestParseVersion(t *testing.T) {
	version, err := ParseVersion("23.255.1")
	if err != nil {
		t.Error(err)
	}
	if uint8(23) != version.Major {
		t.Errorf("Major numbers should match when parsed")
	}
	if uint8(255) != version.Minor {
		t.Errorf("Minor numbers should match when parsed")

	}
	if uint8(1) != version.Patch {
		t.Errorf("Patch numbers should match when parsed")
	}

	_, err = ParseVersion("2312.3.1")
	if err == nil {
		t.Errorf("Major number 2312 is larger than byte")
	}
	_, err = ParseVersion("231.256.1")
	if err == nil {
		t.Errorf("Minor number 256 is larger than byte")
	}
	_, err = ParseVersion("231.3.5645")
	if err == nil {
		t.Errorf("Patch number 5645 is larger than byte")
	}
}
