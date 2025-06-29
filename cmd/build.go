package cmd

import "fmt"

type BuildCmd struct{}

func (b *BuildCmd) Run() error {
	fmt.Println("Hello, World!")
	return nil
}
