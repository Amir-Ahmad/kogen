package generator

import (
	"fmt"
	"io"
	"iter"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type InitGenerator = func(manifest Manifest) (Generator, error)

type Generator interface {
	Generate() (iter.Seq2[Object, error], error)
}

type Object interface {
	GetKind() string
	Output(w io.Writer) error
}

var (
	mu         sync.RWMutex
	generators = map[schema.GroupVersionKind]InitGenerator{}
)

// Register registers a generator for a specific GVK.
func Register(gvk schema.GroupVersionKind, g InitGenerator) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := generators[gvk]; exists {
		panic(fmt.Sprintf("generator already registered for GVK: %v", gvk))
	}
	generators[gvk] = g
}

// GetGenerator returns the generator for a specific GVK.
func GetGenerator(manifest Manifest) (Generator, error) {
	gvk := manifest.GroupVersionKind()
	initFunc, ok := generators[gvk]
	if !ok {
		return nil, fmt.Errorf("generator for %v not found", gvk)
	}
	return initFunc(manifest)
}
