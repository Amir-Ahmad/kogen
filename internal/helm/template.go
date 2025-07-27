package helm

import (
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

// Release represents a helm Release.
type Release struct {
	Release     string                 `json:"release"`
	Namespace   string                 `json:"namespace"`
	IncludeCRDs bool                   `json:"includeCRDs"`
	Values      map[string]interface{} `json:"values"`
	KubeVersion string                 `json:"kubeVersion"`
	APIVersions []string               `json:"apiVersions"`
}

// Template does the equivalent of a `helm template`
func (r Release) Template(chartPath string, kubeVersion string) (map[string]string, error) {
	chart, err := loader.Load(chartPath)
	if err != nil {
		return map[string]string{}, fmt.Errorf("loading chart: %w", err)
	}

	// Use the same default capabilities as helm when not specified
	capabilities := chartutil.DefaultCapabilities

	if r.KubeVersion != "" {
		parsedKubeVersion, err := chartutil.ParseKubeVersion(kubeVersion)
		if err != nil {
			return map[string]string{}, fmt.Errorf("error parsing kubeVersion: %w", err)
		}
		capabilities.KubeVersion = *parsedKubeVersion
	}

	if len(r.APIVersions) > 0 {
		capabilities.APIVersions = r.APIVersions
	}

	if err := chartutil.ProcessDependenciesWithMerge(chart, r.Values); err != nil {
		return map[string]string{}, fmt.Errorf("processing chart dependencies: %w", err)
	}

	releaseOptions := chartutil.ReleaseOptions{
		Name:      r.Release,
		Namespace: r.Namespace,
		IsInstall: true,
	}

	// Merge chart values with our provided ones
	renderValues, err := chartutil.ToRenderValues(chart, r.Values, releaseOptions, capabilities)
	if err != nil {
		return map[string]string{}, fmt.Errorf("preparing render values: %w", err)
	}

	// Run Helm template
	engine := engine.Engine{}
	renderedTemplates, err := engine.Render(chart, renderValues)
	if err != nil {
		return map[string]string{}, fmt.Errorf("rendering helm templates: %w", err)
	}

	// Iterate through charts CRDs and add to map
	if r.IncludeCRDs {
		for _, crd := range chart.CRDObjects() {
			renderedTemplates["crds/"+crd.File.Name] = string(crd.File.Data)
		}
	}

	for key, val := range renderedTemplates {
		// Remove NOTES.txt which is sometimes in the templated output
		if strings.HasSuffix(key, "NOTES.txt") {
			delete(renderedTemplates, key)
		}

		// Delete all empty templates
		if strings.TrimSpace(val) == "" {
			delete(renderedTemplates, key)
		}
	}

	return renderedTemplates, nil
}
