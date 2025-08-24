package version

import "testing"

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
	if Name == "" {
		t.Error("Name should not be empty")
	}
}
