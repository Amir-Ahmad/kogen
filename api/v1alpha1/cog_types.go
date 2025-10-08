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

	// Options specific to helm charts.
	HelmOptions HelmOptions `json:"helmOptions,omitempty"`

	// Specify any kustomization patches or transformers.
	Kustomize kustomize_types.Kustomization `json:"kustomize,omitempty"`
}

type HelmChart struct {
	// Helm release name.
	ReleaseName string `json:"releaseName,omitempty"`

	// Helm chart repository URL. In the case of OCI, it should start with oci://
	Repository string `json:"repository,omitempty"`

	// Helm chart name.
	ChartName string `json:"chartName,omitempty"`

	// Kubernetes namespace to install the chart into.
	Namespace string `json:"namespace,omitempty"`

	// Chart version
	Version string `json:"version,omitempty"`

	// Values to provide to the chart when rendering.
	Values map[string]interface{} `json:"values,omitempty"`

	// equivalent of helm template --include-crds.
	IncludeCRDs bool `json:"includeCRDs,omitempty"`

	// This will create the namespace if it does not exist.
	CreateNamespace bool `json:"createNamespace,omitempty"`
}

type HelmOptions struct {
	KubeVersion string   `json:"kubeVersion,omitempty"`
	APIVersions []string `json:"apiVersions,omitempty"`
}
