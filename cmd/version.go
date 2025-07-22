package cmd

import "fmt"

var (
	version = "v0.0.0-dev"
	commit  = "none"
	date    = "1970-01-01T00:00:00Z"
)

type VersionCmd struct{}

func (cmd *VersionCmd) Run() error {
	fmt.Printf("kogen %s, commit %s, built at %s\n", version, commit, date)
	return nil
}
