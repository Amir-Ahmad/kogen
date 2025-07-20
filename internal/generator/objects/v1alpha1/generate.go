package v1alpha1

import (
	"fmt"
	"io"
	"iter"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"github.com/amir-ahmad/kogen/internal/generator"
)

// Generator implements generator.Generator.
type Generator struct {
	objects cue.Value
}

// Object implements generator.Object.
type Object struct {
	kind  string
	value cue.Value
}

// Compile time check to ensure Generator implements generator.Generator.
var _ generator.Generator = (*Generator)(nil)

// Compile time check to ensure Object implements generator.Object.
var _ generator.Object = (*Object)(nil)

func (o Object) GetKind() string {
	return o.kind
}

func (o Object) Output(w io.Writer) error {
	yamlBytes, err := yaml.Encode(o.value)
	if err != nil {
		return fmt.Errorf("failed to encode object to yaml: %w", err)
	}

	_, err = w.Write(yamlBytes)
	return err
}

func NewGenerator(manifest generator.Manifest) (generator.Generator, error) {
	objects := manifest.Spec.LookupPath(cue.MakePath(cue.Str("objects")))
	if err := objects.Err(); err != nil {
		return nil, fmt.Errorf("failed to lookup objects: %w", err)
	}

	return &Generator{
		objects: objects,
	}, nil
}

// Generate implements generator.Generator.
func (g *Generator) Generate() (iter.Seq2[generator.Object, error], error) {
	iter, err := g.objects.List()
	if err != nil {
		return nil, fmt.Errorf("failed to iterate objects: %w", err)
	}

	return func(yield func(generator.Object, error) bool) {
		for iter.Next() {
			v := iter.Value()
			if err := v.Err(); err != nil {
				yield(Object{}, fmt.Errorf("error getting cue value for object: %w", err))
				return
			}

			kindVal := v.LookupPath(cue.MakePath(cue.Str("kind")))
			if err := v.Err(); err != nil {
				yield(Object{}, fmt.Errorf("failed to get kind for object: %w", err))
				return
			}

			kind, err := kindVal.String()
			if err != nil {
				yield(Object{}, fmt.Errorf("when getting kind as string: %w", err))
				return
			}

			object := Object{
				kind:  kind,
				value: v,
			}

			if !yield(object, nil) {
				return
			}
		}
	}, nil
}
