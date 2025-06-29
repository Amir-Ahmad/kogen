package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kustomize_types "sigs.k8s.io/kustomize/api/types"
)

type Cog struct {
	metav1.TypeMeta `json:",inline"`
	Spec            CogSpec `json:"spec"`
}

type CogSpec struct {
	// A resource references a yaml file containing kubernetes manifests.
	// Each resource can be a file, directory, or a URL.
	Resource []string `json:"resource,omitempty"`
	// Helm charts to render
	Helm []HelmChart `json:"helm,omitempty"`
	// Specify any kustomization patches or transformers
	Kustomize kustomize_types.Kustomization `json:"kustomize,omitempty"`
}

type HelmChart struct {
	// Helm release name
	ReleaseName string `json:"releaseName,omitempty"`
	Repository  string `json:"repository,omitempty"`
	ChartName   string `json:"chartName,omitempty"`
	// Chart version
	Version string                 `json:"version,omitempty"`
	Values  map[string]interface{} `json:"values,omitempty"`
	// equivalent of helm template --include-crds
	IncludeCRDs bool `json:"includeCRDs,omitempty"`
}
