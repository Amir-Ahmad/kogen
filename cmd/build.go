package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/amir-ahmad/kogen/internal/build"
	"github.com/amir-ahmad/kogen/internal/generator"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

type BuildCmd struct {
	Chdir      string   `short:"c" help:"Change directory before running" env:"KOGEN_CHDIR,ARGOCD_ENV_CHDIR"`
	Path       string   `arg:"" name:"path" help:"Cue path to read generator config from" required:"" env:"KOGEN_PATH,ARGOCD_ENV_PATH"`
	Tag        []string `short:"t" help:"Tags to pass to Cue" env:"KOGEN_TAG,ARGOCD_ENV_TAG"`
	Package    string   `help:"Package to load in Cue" env:"KOGEN_PACKAGE,ARGOCD_ENV_PACKAGE"`
	CacheDir   string   `help:"Path to store downloaded artifacts such as helm charts" default:"${cache_dir}" env:"KOGEN_CACHE_DIR"`
	KindFilter string   `short:"k" help:"Regular expression to filter objects by Kind. This is case insensitive and anchored with ^$." env:"KOGEN_KIND_FILTER,ARGOCD_ENV_KIND_FILTER"`
}

func (b *BuildCmd) Run() error {
	if b.Chdir != "" {
		if err := os.Chdir(b.Chdir); err != nil {
			return fmt.Errorf("failed to change directory: %w", err)
		}
	}

	genInputs, err := b.readGeneratorConfig(b.Path)
	if err != nil {
		return err
	}

	options := build.BuildOptions{
		CacheDir: b.CacheDir,
	}

	if b.KindFilter != "" {
		kindFilter, err := regexp.Compile(fmt.Sprintf("(?i)^%s$", b.KindFilter))
		if err != nil {
			return fmt.Errorf("failed to compile kind filter: %w", err)
		}
		options.KindFilter = kindFilter
	}

	return build.Run(os.Stdout, genInputs, options)
}

func (b *BuildCmd) readGeneratorConfig(loadPath string) ([]generator.GeneratorInput, error) {
	ctx := cuecontext.New()
	cfg := load.Config{Tags: b.Tag}
	if b.Package != "" {
		cfg.Package = b.Package
	}

	genInputs := []generator.GeneratorInput{}

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
			return nil, fmt.Errorf("couldn't find generator config: %w", err)
		}

		iter, err := kogenValue.Fields()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate generator config: %w", err)
		}

		for iter.Next() {
			label := iter.Selector()
			v := iter.Value()
			if err := v.Err(); err != nil {
				return nil, fmt.Errorf("error getting cue value for %s: %w", label, err)
			}

			var genInput generator.GeneratorInput

			if err := v.Decode(&genInput); err != nil {
				return nil, fmt.Errorf("failed to decode generator config for %s: %w", label, err)
			}

			genInput.InstanceDir = inst.Dir
			genInputs = append(genInputs, genInput)
		}
	}
	return genInputs, nil
}
