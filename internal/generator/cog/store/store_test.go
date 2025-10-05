package store

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetObjectKey(t *testing.T) {
	tests := map[string]struct {
		obj         *Object
		expectedKey string
	}{
		"basic object": {
			obj: &Object{
				Unstructured: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "test-cm",
							"namespace": "default",
						},
					},
				},
			},
			expectedKey: "v1|ConfigMap|default|test-cm",
		},
		"cluster-scoped object": {
			obj: &Object{
				Unstructured: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Namespace",
						"metadata": map[string]interface{}{
							"name": "test-ns",
						},
					},
				},
			},
			expectedKey: "v1|Namespace||test-ns",
		},
		"custom resource": {
			obj: &Object{
				Unstructured: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "custom.io/v1alpha1",
						"kind":       "CustomResource",
						"metadata": map[string]interface{}{
							"name":      "my-resource",
							"namespace": "prod",
						},
					},
				},
			},
			expectedKey: "custom.io/v1alpha1|CustomResource|prod|my-resource",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			key := getObjectKey(tc.obj)
			assert.Equal(t, tc.expectedKey, key)
		})
	}
}

func TestObjectStore_Add(t *testing.T) {
	tests := map[string]struct {
		addObjects  []*Object
		expectError bool
		errorMsg    string
	}{
		"add single object": {
			addObjects: []*Object{
				{
					Unstructured: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "test-cm",
								"namespace": "default",
							},
						},
					},
				},
			},
			expectError: false,
		},
		"add multiple different objects": {
			addObjects: []*Object{
				{
					Unstructured: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "cm-1",
								"namespace": "default",
							},
						},
					},
				},
				{
					Unstructured: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "cm-2",
								"namespace": "default",
							},
						},
					},
				},
			},
			expectError: false,
		},
		"add duplicate object": {
			addObjects: []*Object{
				{
					Unstructured: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "test-cm",
								"namespace": "default",
							},
						},
					},
				},
				{
					Unstructured: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "test-cm",
								"namespace": "default",
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "already exists",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			store := NewObjectStore()
			var err error

			for _, obj := range tc.addObjects {
				err = store.Add(obj)
				if err != nil {
					break
				}
			}

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Len(t, *store, len(tc.addObjects))
			}
		})
	}
}

func TestObjectStore_AddYaml(t *testing.T) {
	tests := map[string]struct {
		yaml          string
		expectedCount int
		expectError   bool
	}{
		"single document": {
			yaml: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: default
data:
  key: value
`,
			expectedCount: 1,
			expectError:   false,
		},
		"multiple documents": {
			yaml: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-1
  namespace: default
---
apiVersion: v1
kind: Secret
metadata:
  name: secret-1
  namespace: default
type: Opaque
`,
			expectedCount: 2,
			expectError:   false,
		},
		"empty yaml": {
			yaml:          "",
			expectedCount: 0,
			expectError:   false,
		},
		"yaml with comments": {
			yaml: `
# This is a comment
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: default
`,
			expectedCount: 1,
			expectError:   false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			store := NewObjectStore()
			err := store.AddYaml([]byte(tc.yaml))

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, *store, tc.expectedCount)
			}
		})
	}
}

func TestObjectStore_GetIterator(t *testing.T) {
	tests := map[string]struct {
		objects       []*Object
		expectedOrder []string // expected names in order
	}{
		"empty store": {
			objects:       []*Object{},
			expectedOrder: nil,
		},
		"single object": {
			objects: []*Object{
				{
					Unstructured: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "test-cm",
								"namespace": "default",
							},
						},
					},
				},
			},
			expectedOrder: []string{"test-cm"},
		},
		"sorted by key": {
			objects: []*Object{
				{
					Unstructured: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "z-cm",
								"namespace": "default",
							},
						},
					},
				},
				{
					Unstructured: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "a-cm",
								"namespace": "default",
							},
						},
					},
				},
				{
					Unstructured: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "m-cm",
								"namespace": "default",
							},
						},
					},
				},
			},
			expectedOrder: []string{"a-cm", "m-cm", "z-cm"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			store := NewObjectStore()
			for _, obj := range tc.objects {
				err := store.Add(obj)
				require.NoError(t, err)
			}

			var names []string
			for obj, err := range store.GetIterator() {
				require.NoError(t, err)
				names = append(names, obj.(*Object).GetName())
			}

			assert.Equal(t, tc.expectedOrder, names)
		})
	}
}

func TestObject_Output(t *testing.T) {
	tests := map[string]struct {
		obj              *Object
		expectedContains []string
	}{
		"configmap output": {
			obj: &Object{
				Unstructured: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "test-cm",
							"namespace": "default",
						},
						"data": map[string]interface{}{
							"key": "value",
						},
					},
				},
			},
			expectedContains: []string{
				"apiVersion: v1",
				"kind: ConfigMap",
				"name: test-cm",
				"namespace: default",
				"key: value",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tc.obj.Output(&buf)
			require.NoError(t, err)

			output := buf.String()
			for _, expected := range tc.expectedContains {
				assert.Contains(t, output, expected)
			}

			// Verify it's valid YAML
			assert.True(t, strings.HasPrefix(strings.TrimSpace(output), "apiVersion:") ||
				strings.Contains(output, "apiVersion:"))
		})
	}
}

func TestObject_GetKind(t *testing.T) {
	tests := map[string]struct {
		obj          *Object
		expectedKind string
	}{
		"configmap": {
			obj: &Object{
				Unstructured: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "test",
						},
					},
				},
			},
			expectedKind: "ConfigMap",
		},
		"deployment": {
			obj: &Object{
				Unstructured: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "apps/v1",
						"kind":       "Deployment",
						"metadata": map[string]interface{}{
							"name": "test",
						},
					},
				},
			},
			expectedKind: "Deployment",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			kind := tc.obj.GetKind()
			assert.Equal(t, tc.expectedKind, kind)
		})
	}
}
