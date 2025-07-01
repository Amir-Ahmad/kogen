package build

import (
	"io"

	"github.com/amir-ahmad/kogen/api/v1alpha1"
	cog_v1alpha1 "github.com/amir-ahmad/kogen/internal/generator/cog/v1alpha1"
	"github.com/amir-ahmad/kogen/internal/registry"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	registry.Register(v1alpha1.CogGVK, cog_v1alpha1.NewGenerator)
}

func Generate(w io.Writer, manifests []unstructured.Unstructured) error {
	for _, manifest := range manifests {
		gen, err := registry.GetGenerator(manifest)
		if err != nil {
			return err
		}
		st, err := gen.Generate()
		if err != nil {
			return err
		}

		if err := st.Output(w); err != nil {
			return err
		}
	}
	return nil
}
