package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/amir-ahmad/kogen/internal/build"
	"github.com/amir-ahmad/kogen/internal/generator"
	"github.com/amir-ahmad/kogen/internal/sops"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
)

type BuildCmd struct {
	// flags with short options
	Chdir      string   `short:"c" help:"Change directory before running"                                                              env:"KOGEN_CHDIR,ARGOCD_ENV_CHDIR"`
	KindFilter string   `short:"k" help:"Regular expression to filter objects by Kind. This is case insensitive and anchored with ^$." env:"KOGEN_KIND_FILTER,ARGOCD_ENV_KIND_FILTER"`
	Package    string   `short:"p" help:"Package to load in Cue"                                                                       env:"KOGEN_PACKAGE,ARGOCD_ENV_PACKAGE"`
	Tag        []string `short:"t" help:"Tags to pass to Cue"                                                                          env:"KOGEN_TAG,ARGOCD_ENV_TAG"`

	// flags without short options
	CacheDir   string `help:"Path to store downloaded artifacts such as helm charts"                    env:"KOGEN_CACHE_DIR,ARGOCD_ENV_KOGEN_CACHE_DIR"   default:"${cache_dir}"`
	KogenField string `help:"Top level field to find kogen components. Defaults to kogen by convention" env:"KOGEN_FIELD,ARGOCD_ENV_KOGEN_FIELD"           default:"kogen"`
	SopsField  string `help:"Top level field to recursively find sops attribute and decode."            env:"KOGEN_SOPS_FIELD,ARGOCD_ENV_KOGEN_SOPS_FIELD" default:"secrets"`

	// positional args
	Path string `arg:"" name:"path" help:"Cue path to read generator config from" required:"" env:"KOGEN_PATH,ARGOCD_ENV_KOGEN_PATH"`
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

		instanceValue, err := sops.Inject(
			instanceValue,
			cue.ParsePath(b.SopsField),
			inst.Dir,
			inst.Root,
		)
		if err != nil {
			return nil, err
		}

		kogenValue := instanceValue.LookupPath(cue.ParsePath(b.KogenField))
		if err := kogenValue.Err(); err != nil {
			return nil, fmt.Errorf("couldn't find generator config: %w", err)
		}

		if err := kogenValue.Validate(cue.Concrete(true)); err != nil {
			return nil, fmt.Errorf("when validating cue: %w", formatCueError(err))
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

func formatCueError(err error) error {
	errs := errors.Errors(err)

	return errors.New(fmt.Sprintf(
		"%v\n\n# Error details (%d):\n%v\n",
		err,
		len(errs),
		errors.Details(err, nil),
	))
}
