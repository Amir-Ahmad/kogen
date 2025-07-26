package generator

import (
	"fmt"
	"io"
	"iter"
	"sync"

	"cuelang.org/go/cue"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	mu         sync.RWMutex
	generators = map[schema.GroupVersionKind]InitGenerator{}
)

// InitGenerator is a function to initialize a generator from its manifest.
type InitGenerator = func(manifest Manifest) (Generator, error)

// All generators must implement this interface.
type Generator interface {
	Generate() (iter.Seq2[Object, error], error)
}

type Object interface {
	// GetName returns the name of the object.
	GetKind() string
	// Output writes the object to the provided writer in yaml format.
	Output(w io.Writer) error
}

// Manifest represents the manifest that contains the generator's configuration.
type Manifest struct {
	metav1.TypeMeta `json:",inline"`
	Spec            cue.Value
}

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
