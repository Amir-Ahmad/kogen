package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amir-ahmad/kogen/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"kogen": cmd.Execute,
	})
}

// TestScripts runs the tests in the testdata/testscripts directory
func TestScripts(t *testing.T) {
	modCacheDir := cacheDir(t)

	testscript.Run(t, testscript.Params{
		Setup: setupTestEnvironment(modCacheDir),
		Dir:   "testdata/testscripts",
		// Set UPDATE_GOLDEN to 0 to update golden files
		UpdateScripts:       os.Getenv("UPDATE_GOLDEN") == "0",
		RequireExplicitExec: true,
		Condition: func(cond string) (bool, error) {
			// Set RUN_REMOTE to 0 to run remote tests
			if cond == "remote" {
				return os.Getenv("RUN_REMOTE") == "0", nil
			}
			return false, nil
		},
	})
}

// cacheDir returns the cue cache directory, this will be used to cache modules in tests
func cacheDir(t *testing.T) string {
	t.Helper()

	if dir := os.Getenv("CUE_CACHE_DIR"); dir != "" {
		return dir
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		t.Fatalf("failed to get system cache directory: %v", err)
	}
	return filepath.Join(dir, "cue")
}

func setupTestEnvironment(cacheDir string) func(*testscript.Env) error {
	return func(env *testscript.Env) error {
		// Reuse cache directory and use my
		env.Vars = append(env.Vars,
			"CUE_REGISTRY=github.com/amir-ahmad=ghcr.io",
			"CUE_CACHE_DIR="+cacheDir,
		)
		return nil
	}
}
