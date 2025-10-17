package v1alpha1

import (
	"crypto/sha256"
	"fmt"
	"io"
	"iter"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/amir-ahmad/kogen/api/v1alpha1"
	"github.com/amir-ahmad/kogen/internal/generator"
	"github.com/amir-ahmad/kogen/internal/generator/cog/store"
	"github.com/amir-ahmad/kogen/internal/helm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Generator implements generator.Generator.
type Generator struct {
	spec        v1alpha1.CogSpec
	instanceDir string
}

// Compile time check to ensure Generator implements generator.Generator.
var _ generator.Generator = (*Generator)(nil)

func NewGenerator(input generator.GeneratorInput) (generator.Generator, error) {
	var spec v1alpha1.CogSpec
	if err := input.Spec.Decode(&spec); err != nil {
		return nil, fmt.Errorf("when decoding cog spec: %w", err)
	}

	return &Generator{
		spec:        spec,
		instanceDir: input.InstanceDir,
	}, nil
}

func isZero[T any](v T) bool {
	return reflect.ValueOf(v).IsZero()
}

// Generate implements generator.Generator.
func (g *Generator) Generate(
	options generator.Options,
) (iter.Seq2[generator.Object, error], error) {
	st := store.NewObjectStore()

	for _, resource := range g.spec.Resource {
		if err := addResourceObjects(st, resource, g.instanceDir, filepath.Join(options.CacheDir, "resources")); err != nil {
			return nil, err
		}
	}

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

	var err error
	// Replace store with kustomize.
	if !isZero(g.spec.Kustomize) {
		st, err = processStoreWithKustomize(st, g.spec.Kustomize)
		if err != nil {
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

	// Create namespace if requested
	if helmChart.CreateNamespace && helmChart.Namespace != "" {
		namespace := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": helmChart.Namespace,
				},
			},
		}
		if err := st.Add(&store.Object{Unstructured: namespace}); err != nil {
			return fmt.Errorf("when adding namespace object to store: %w", err)
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

	for k, v := range renderedTemplates {
		if err := st.AddYaml([]byte(v)); err != nil {
			return fmt.Errorf("when adding helm objects to store from %s: %w", k, err)
		}
	}

	return nil
}

// addResourceObjects adds all objects from a resource to the store.
func addResourceObjects(
	st *store.ObjectStore,
	resource string,
	instanceDir string,
	cacheDir string,
) error {
	var yamlData []byte
	var err error

	if strings.HasPrefix(resource, "http://") || strings.HasPrefix(resource, "https://") {
		// Handle http yaml files with caching
		yamlData, err = getHTTPResource(resource, cacheDir)
		if err != nil {
			return fmt.Errorf("when getting cached resource from URL %s: %w", resource, err)
		}
	} else {
		// Handle local file/directory resources
		resourcePath := resource
		if !filepath.IsAbs(resource) {
			resourcePath = filepath.Join(instanceDir, resource)
		}

		info, err := os.Stat(resourcePath)
		if err != nil {
			return fmt.Errorf("when accessing resource %s: %w", resourcePath, err)
		}

		if info.IsDir() {
			// Handle directory - read all yaml files
			return addResourceDirectory(st, resourcePath)
		} else {
			// Handle single file
			yamlData, err = os.ReadFile(resourcePath)
			if err != nil {
				return fmt.Errorf("when reading resource file %s: %w", resourcePath, err)
			}
		}
	}

	if err := st.AddYaml(yamlData); err != nil {
		return fmt.Errorf("when adding resource objects to store: %w", err)
	}

	return nil
}

// getHTTPResource fetches a resource from a remote URL and caches it.
func getHTTPResource(url string, cacheDir string) ([]byte, error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("when creating cache directory: %w", err)
	}

	// Generate cache key from URL, with first 8 bytes.
	hash := sha256.Sum256([]byte(url))
	cacheKey := fmt.Sprintf("%x", hash[:8])
	cacheFile := filepath.Join(cacheDir, cacheKey+".yaml")

	// Read from cached file when it exists
	if _, err := os.Stat(cacheFile); err == nil {
		data, err := os.ReadFile(cacheFile)
		if err != nil {
			return nil, fmt.Errorf("when reading cached resource: %w", err)
		}
		return data, nil
	}

	// Fetch from remote if not cached
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("when fetching resource from URL %s: %w", url, err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to fetch resource from URL %s: status %d",
			url,
			resp.StatusCode,
		)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("when reading resource from URL %s: %w", url, err)
	}

	if err := os.WriteFile(cacheFile, data, 0o644); err != nil {
		return nil, fmt.Errorf("when caching resource from URL %s: %w", url, err)
	}

	return data, nil
}

// addResourceDirectory adds all yaml files in a directory to the store.
func addResourceDirectory(st *store.ObjectStore, dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("unable to read directory %s: %w", dirPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		fullPath := filepath.Join(dirPath, entry.Name())
		yamlData, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("when reading resource file %s: %w", fullPath, err)
		}

		if err := st.AddYaml(yamlData); err != nil {
			return fmt.Errorf("when adding resource objects from %s to store: %w", fullPath, err)
		}
	}

	return nil
}
