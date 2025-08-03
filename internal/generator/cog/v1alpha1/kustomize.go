package v1alpha1

import (
	"bytes"
	"fmt"

	"github.com/amir-ahmad/kogen/internal/generator/cog/store"
	"sigs.k8s.io/kustomize/api/krusty"
	kustomize_types "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/yaml"
)

// processStoreWithKustomize applies kustomization to store objects and returns
// a new store with the processed results.
func processStoreWithKustomize(
	st *store.ObjectStore,
	kustomization kustomize_types.Kustomization,
) (*store.ObjectStore, error) {
	memfs := filesys.MakeFsInMemory()

	// Write all store objects to memfs.
	byteBuffer := &bytes.Buffer{}
	for object := range st.GetIterator() {
		if err := object.Output(byteBuffer); err != nil {
			return nil, err
		}

		fmt.Fprintf(byteBuffer, "---\n")
	}
	if err := memfs.WriteFile("cog_objects.yaml", byteBuffer.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to write store to kustomize memfs: %w", err)
	}

	kustomization.Resources = append(kustomization.Resources, "cog_objects.yaml")

	// Write kustomization.yaml to memfs.
	kustBytes, err := yaml.Marshal(kustomization)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal kustomization: %w", err)
	}

	if err := memfs.WriteFile("kustomization.yaml", kustBytes); err != nil {
		return nil, fmt.Errorf("failed to write kustomization.yaml to kustomize memfs: %w", err)
	}

	kustomizer := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	resmap, err := kustomizer.Run(memfs, ".")
	if err != nil {
		return nil, fmt.Errorf("when running kustomize: %w", err)
	}

	kustOutput, err := resmap.AsYaml()
	if err != nil {
		return nil, fmt.Errorf("when marshaling kustomize output: %w", err)
	}

	newStore := store.NewObjectStore()
	if err := newStore.AddYaml(kustOutput); err != nil {
		return nil, fmt.Errorf("when adding kustomize output to store: %w", err)
	}

	return newStore, nil
}
