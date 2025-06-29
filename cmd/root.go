package cmd

import (
	"github.com/alecthomas/kong"
)

type Cli struct {
	Build BuildCmd `cmd:"" help:"Generate Kubernetes manifests"`
}

func Execute() {
	cli := Cli{}
	ctx := kong.Parse(&cli)
	err := ctx.Run(&cli)
	ctx.FatalIfErrorf(err)
}
