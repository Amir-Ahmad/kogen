package v1alpha1

import (
	"fmt"
	"iter"
	"path/filepath"

	"github.com/amir-ahmad/kogen/api/v1alpha1"
	"github.com/amir-ahmad/kogen/internal/generator"
	"github.com/amir-ahmad/kogen/internal/generator/cog/store"
	"github.com/amir-ahmad/kogen/internal/helm"
)

// Generator implements generator.Generator.
type Generator struct {
	spec        v1alpha1.CogSpec
	instanceDir string
}

// Compile time check to ensure Generator implements generator.Generator.
var _ generator.Generator = (*Generator)(nil)

func NewGenerator(manifest generator.Manifest) (generator.Generator, error) {
	var spec v1alpha1.CogSpec
	if err := manifest.Spec.Decode(&spec); err != nil {
		return nil, fmt.Errorf("when decoding cog spec: %w", err)
	}

	return &Generator{
		spec:        spec,
		instanceDir: manifest.InstanceDir,
	}, nil
}

// Generate implements generator.Generator.
func (g *Generator) Generate(options generator.Options) (iter.Seq2[generator.Object, error], error) {
	st := store.NewObjectStore()

	for _, h := range g.spec.Helm {
		if err := addHelmObjects(
			st,
			h,
			g.spec.HelmOptions,
			filepath.Join(options.CacheDir, "helm"),
			g.instanceDir,
		); err != nil {
			return nil, err
		}
	}

	return st.GetIterator(), nil
}

func addHelmObjects(
	st *store.ObjectStore,
	helmChart v1alpha1.HelmChart,
	helmOptions v1alpha1.HelmOptions,
	cacheDir string,
	instanceDir string,
) error {
	chart := helm.Chart{
		Repository: helmChart.Repository,
		ChartName:  helmChart.ChartName,
		Version:    helmChart.Version,
	}

	chartDir := filepath.Join(instanceDir, helmChart.Repository)
	var err error
	if chart.GetChartType() != helm.ChartTypeLocal {
		chartDir, err = chart.DownloadChart(cacheDir)
		if err != nil {
			return fmt.Errorf("when downloading chart: %w", err)
		}
	}

	release := helm.Release{
		Release:     helmChart.ReleaseName,
		Namespace:   helmChart.Namespace,
		IncludeCRDs: helmChart.IncludeCRDs,
		Values:      helmChart.Values,
		KubeVersion: helmOptions.KubeVersion,
		APIVersions: helmOptions.APIVersions,
	}

	renderedTemplates, err := release.Template(chartDir)
	if err != nil {
		return fmt.Errorf("when rendering helm templates: %w", err)
	}

	for _, v := range renderedTemplates {
		if err := st.AddYaml([]byte(v)); err != nil {
			return fmt.Errorf("when adding helm objects to store: %w", err)
		}
	}

	return nil
}
