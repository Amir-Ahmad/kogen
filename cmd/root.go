package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
)

type Cli struct {
	Build   BuildCmd   `cmd:"" help:"Generate Kubernetes manifests"`
	Version VersionCmd `cmd:"" help:"Show version information"`
}

func getCacheDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, "kogen"), err
}

func Execute() {
	cli := Cli{}
	cacheDir, err := getCacheDir()
	if err != nil {
		fmt.Printf("failed to get cache directory: %v\n", err)
		os.Exit(1)
	}
	ctx := kong.Parse(&cli, kong.Vars{
		"cache_dir": cacheDir,
	})
	err = ctx.Run(&cli)
	ctx.FatalIfErrorf(err)
}
