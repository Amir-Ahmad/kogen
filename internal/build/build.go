package build

import (
	"fmt"
	"io"
	"regexp"

	"github.com/amir-ahmad/kogen/api/v1alpha1"
	"github.com/amir-ahmad/kogen/internal/generator"
	cog_v1alpha1 "github.com/amir-ahmad/kogen/internal/generator/cog/v1alpha1"
	obj_v1alpha1 "github.com/amir-ahmad/kogen/internal/generator/objects/v1alpha1"
)

// BuildOptions are options to configure kogen build.
type BuildOptions struct {
	// CacheDir is the directory to use for downloading artifacts.
	CacheDir string

	// KindFilter is a regular expression to filter objects by Kind.
	KindFilter *regexp.Regexp
}

func init() {
	// Register GVKs with their init functions.
	generator.Register(v1alpha1.CogGVK, cog_v1alpha1.NewGenerator)
	generator.Register(v1alpha1.ObjectsGVK, obj_v1alpha1.NewGenerator)
}

func Run(w io.Writer, genInputs []generator.GeneratorInput, opts BuildOptions) error {
	genOptions := generator.Options{
		CacheDir: opts.CacheDir,
	}

	printSeparator := false
	for _, genInput := range genInputs {
		gen, err := generator.GetGenerator(genInput)
		if err != nil {
			return err
		}

		it, err := gen.Generate(genOptions)
		if err != nil {
			return err
		}

		for object, err := range it {
			if err != nil {
				return err
			}

			if opts.KindFilter != nil && !opts.KindFilter.MatchString(object.GetKind()) {
				continue
			}

			// The separator needs to be printed after every object but the last.
			if printSeparator {
				fmt.Fprintf(w, "---\n")
			} else {
				printSeparator = true
			}

			if err := object.Output(w); err != nil {
				return err
			}
		}
	}
	return nil
}
