package test

import (
	"os"
	"testing"

	"github.com/amir-ahmad/kogen/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"kogen": cmd.Execute,
	})
}

// TestScripts runs the tests in the testdata/testscripts directory.
func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/testscripts",
		// Set UPDATE_GOLDEN to 0 to update golden files.
		UpdateScripts:       os.Getenv("UPDATE_GOLDEN") == "0",
		RequireExplicitExec: true,
		Condition: func(cond string) (bool, error) {
			// Set RUN_REMOTE to 0 to run remote tests.
			if cond == "remote" {
				return os.Getenv("RUN_REMOTE") == "0", nil
			}
			return false, nil
		},
	})
}
