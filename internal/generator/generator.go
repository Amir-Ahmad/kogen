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

// InitGenerator is a function to initialize a generator from its input/spec.
type InitGenerator = func(input GeneratorInput) (Generator, error)

// All generators must implement this interface.
type Generator interface {
	Generate(options Options) (iter.Seq2[Object, error], error)
}

// Generators return an iterator of Objects.
type Object interface {
	// GetName returns the name of the object.
	GetKind() string
	// Output writes the object to the provided writer in yaml format.
	Output(w io.Writer) error
}

// Options for generators.
type Options struct {
	// CacheDir is the directory to use for downloading artifacts.
	CacheDir string
}

// GeneratorInput is the input to a generator.
type GeneratorInput struct {
	metav1.TypeMeta `json:",inline"`
	Spec            cue.Value

	// InstanceDir is the directory that the config was loaded from.
	InstanceDir string
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
func GetGenerator(input GeneratorInput) (Generator, error) {
	gvk := input.GroupVersionKind()
	initFunc, ok := generators[gvk]
	if !ok {
		return nil, fmt.Errorf("generator for %v not found", gvk)
	}
	return initFunc(input)
}
