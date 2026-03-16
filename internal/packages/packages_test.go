package packages

import (
	"testing"
)

func TestDetect_ReturnsManagerOrError(t *testing.T) {
	// We cannot easily mock PATH in a unit test, so we simply verify that
	// Detect returns either a valid PackageManager or an error — never both nil.
	pm, err := Detect()
	if pm == nil && err == nil {
		t.Fatal("Detect returned nil PackageManager and nil error; expected one or the other")
	}
	if pm != nil && err != nil {
		t.Fatalf("Detect returned both a PackageManager and an error: %v", err)
	}
}

func TestRun_EmptyPackageList(t *testing.T) {
	err := Run(nil)
	if err != nil {
		t.Fatalf("Run(nil) returned error: %v", err)
	}

	err = Run([]string{})
	if err != nil {
		t.Fatalf("Run([]string{}) returned error: %v", err)
	}
}
