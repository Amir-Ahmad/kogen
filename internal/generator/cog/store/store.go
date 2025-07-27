package store

import (
	"bytes"
	"fmt"
	"io"
	"iter"

	"github.com/amir-ahmad/kogen/internal/generator"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	util_yaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

type Object struct {
	*unstructured.Unstructured
}

type ObjectStore map[string]*Object

// Compile time check to ensure Object implements generator.Object.
var _ generator.Object = (*Object)(nil)

// Output implements generator.Object.Output.
func (o *Object) Output(w io.Writer) error {
	yamlBytes, err := yaml.Marshal(o.Unstructured)
	if err != nil {
		return fmt.Errorf("failed to encode object to yaml: %w", err)
	}

	_, err = w.Write(yamlBytes)
	return err
}

// NewObjectStore creates a new object store.
func NewObjectStore() *ObjectStore {
	return &ObjectStore{}
}

// Add adds an object to the store.
// It returns an error if an object with the same key is already present.
func (s *ObjectStore) Add(obj *Object) error {
	key := getObjectKey(obj)
	if _, ok := (*s)[key]; ok {
		return fmt.Errorf("object %s already exists", key)
	}
	(*s)[key] = obj
	return nil
}

// AddYaml adds yaml objects to the store.
func (s *ObjectStore) AddYaml(yamlBytes []byte) error {
	decoder := util_yaml.NewYAMLToJSONDecoder(bytes.NewReader(yamlBytes))
	for {
		var manifest unstructured.Unstructured
		if err := decoder.Decode(&manifest); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		s.Add(&Object{&manifest})
	}
	return nil
}

// GetIterator returns an iterator for the objects in the store.
func (s *ObjectStore) GetIterator() (iter.Seq2[generator.Object, error], error) {
	return func(yield func(generator.Object, error) bool) {
		for _, obj := range *s {
			if !yield(obj, nil) {
				return
			}
		}
	}, nil
}

// getObjectKey returns a unique key for an object.
func getObjectKey(obj *Object) string {
	return fmt.Sprintf(
		"%s|%s|%s|%s",
		obj.GroupVersionKind().GroupVersion(),
		obj.GroupVersionKind().Kind,
		obj.GetNamespace(),
		obj.GetName(),
	)
}
