package store

import (
	"fmt"
	"io"
	"maps"
	"slices"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

type Object = unstructured.Unstructured

type Store interface {
	// Add adds an object to the store
	Add(obj *Object) error

	// Objects returns all objects in the store
	Objects() ([]*Object, error)

	// Output writes all objects as yaml to the writer.
	Output(w io.Writer) error
}

type ObjectStore map[string]Object

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
	(*s)[key] = *obj
	return nil
}

// Objects returns all objects in the store.
func (s *ObjectStore) Objects() ([]*Object, error) {
	var objects []*Object
	for _, obj := range *s {
		objects = append(objects, &obj)
	}
	return objects, nil
}

// getObjectKey returns a unique key for an object.
func getObjectKey(obj *Object) string {
	return fmt.Sprintf("%s|%s|%s|%s", obj.GroupVersionKind().GroupVersion(), obj.GroupVersionKind().Kind, obj.GetNamespace(), obj.GetName())
}

// Output writes all objects as yaml to the writer.
func (s *ObjectStore) Output(w io.Writer) error {
	yamlSelector := "---\n"

	keys := slices.Sorted(maps.Keys(*s))

	for _, key := range keys {
		obj := (*s)[key]
		yamlData, err := yaml.Marshal(obj.Object)
		if err != nil {
			return fmt.Errorf("failed to marshal object %s: %w", key, err)
		}
		if _, err := w.Write(yamlData); err != nil {
			return fmt.Errorf("failed to write object %s: %w", key, err)
		}
		if _, err := w.Write([]byte(yamlSelector)); err != nil {
			return fmt.Errorf("failed to write yaml selector: %w", err)
		}
	}
	return nil
}
