package cmd

import (
	"fmt"
	"os"

	"github.com/amir-ahmad/kogen/internal/build"
	"github.com/amir-ahmad/kogen/internal/generator"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

type BuildCmd struct {
	Chdir   string   `short:"c" help:"Change directory before running"`
	Path    string   `short:"p" help:"Cue path to read manifests from" required:""`
	Tag     []string `short:"t" help:"Tags to pass to Cue"`
	Package string   `help:"Package to load in Cue"`
}

func (b *BuildCmd) Run() error {
	if b.Chdir != "" {
		if err := os.Chdir(b.Chdir); err != nil {
			return fmt.Errorf("failed to change directory: %w", err)
		}
	}

	manifests, err := b.readManifests(b.Path)
	if err != nil {
		return err
	}

	return build.Generate(os.Stdout, manifests)
}

func (b *BuildCmd) readManifests(loadPath string) ([]generator.Manifest, error) {
	ctx := cuecontext.New()
	cfg := load.Config{Tags: b.Tag}
	if b.Package != "" {
		cfg.Package = b.Package
	}

	manifests := []generator.Manifest{}

	insts := load.Instances([]string{loadPath}, &cfg)
	for _, inst := range insts {
		if inst.Err != nil {
			return nil, fmt.Errorf("error when loading cue instance: %w", inst.Err)
		}

		instanceValue := ctx.BuildInstance(inst)
		if err := instanceValue.Err(); err != nil {
			return nil, fmt.Errorf("failed to build cue instance: %w", err)
		}

		kogenValue := instanceValue.LookupPath(cue.ParsePath("kogen"))
		if err := kogenValue.Err(); err != nil {
			return nil, fmt.Errorf("couldn't find manifests: %w", err)
		}

		iter, err := kogenValue.Fields()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate manifests: %w", err)
		}

		for iter.Next() {
			label := iter.Selector()
			v := iter.Value()
			if err := v.Err(); err != nil {
				return nil, fmt.Errorf("error getting cue value for %s: %w", label, err)
			}

			var manifest generator.Manifest

			if err := v.Decode(&manifest); err != nil {
				return nil, fmt.Errorf("failed to decode manifest for %s: %w", label, err)
			}

			manifests = append(manifests, manifest)
		}
	}
	return manifests, nil
}
