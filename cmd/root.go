package cmd

import (
	"github.com/alecthomas/kong"
)

type Cli struct {
	Build   BuildCmd   `cmd:"" help:"Generate Kubernetes manifests"`
	Version VersionCmd `cmd:"" help:"Show version information"`
}

func Execute() {
	cli := Cli{}
	ctx := kong.Parse(&cli)
	err := ctx.Run(&cli)
	ctx.FatalIfErrorf(err)
}
