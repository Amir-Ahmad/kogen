package v1alpha1

import (
	"github.com/amir-ahmad/kogen/internal/registry"
	"github.com/amir-ahmad/kogen/internal/store"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type CogGenerator struct{}

// Compile time check to ensure CogGenerator implements registry.Generator.
var _ registry.Generator = (*CogGenerator)(nil)

func NewGenerator(manifest unstructured.Unstructured) (registry.Generator, error) {
	return &CogGenerator{}, nil
}

// Generate implements registry.Generator.
func (g *CogGenerator) Generate() (store.Store, error) {
	return store.NewObjectStore(), nil
}
