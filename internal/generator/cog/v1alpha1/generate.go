package v1alpha1

import (
	"iter"

	"github.com/amir-ahmad/kogen/internal/generator"
)

type CogGenerator struct{}

// Compile time check to ensure CogGenerator implements generator.Generator.
var _ generator.Generator = (*CogGenerator)(nil)

func NewGenerator(manifest generator.Manifest) (generator.Generator, error) {
	return &CogGenerator{}, nil
}

// Generate implements generator.Generator.
func (g *CogGenerator) Generate() (iter.Seq2[generator.Object, error], error) {
	return nil, nil
}
