package build

import (
	"fmt"
	"io"

	"github.com/amir-ahmad/kogen/api/v1alpha1"
	"github.com/amir-ahmad/kogen/internal/generator"
	cog_v1alpha1 "github.com/amir-ahmad/kogen/internal/generator/cog/v1alpha1"
	obj_v1alpha1 "github.com/amir-ahmad/kogen/internal/generator/objects/v1alpha1"
)

func init() {
	generator.Register(v1alpha1.CogGVK, cog_v1alpha1.NewGenerator)
	generator.Register(v1alpha1.ObjectsGVK, obj_v1alpha1.NewGenerator)
}

func Generate(w io.Writer, manifests []generator.Manifest) error {
	for _, manifest := range manifests {
		gen, err := generator.GetGenerator(manifest)
		if err != nil {
			return err
		}
		it, err := gen.Generate()
		if err != nil {
			return err
		}

		for object, err := range it {
			if err != nil {
				return err
			}

			if err := object.Output(w); err != nil {
				return err
			}

			fmt.Fprintf(w, "---\n")
		}
	}
	return nil
}
