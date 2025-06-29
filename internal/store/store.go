package store

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Object = unstructured.Unstructured

type Store interface {
	// Add adds an object to the store
	Add(obj *Object) error

	// Objects returns all objects in the store
	Objects() ([]*Object, error)
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
