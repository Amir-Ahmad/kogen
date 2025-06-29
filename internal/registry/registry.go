package registry

import (
	"fmt"
	"sync"

	"github.com/amir-ahmad/kogen/internal/store"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type InitGenerator = func(manifest unstructured.Unstructured) (Generator, error)

type Generator interface {
	Generate() (store.Store, error)
}

var (
	mu         sync.Mutex
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
func GetGenerator(manifest unstructured.Unstructured) (Generator, error) {
	gvk := manifest.GroupVersionKind()
	initFunc, ok := generators[gvk]
	if !ok {
		return nil, fmt.Errorf("generator for %v not found", gvk)
	}
	return initFunc(manifest)
}
