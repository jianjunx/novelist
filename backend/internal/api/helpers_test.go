package api

import (
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	// Check if CGO is available (needed for go-sqlite3)
	if err := exec.Command("gcc", "--version").Run(); err != nil {
		// No gcc available, skip all tests in this package
		return
	}
	os.Exit(m.Run())
}
