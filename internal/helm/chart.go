package helm

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
)

// Chart represents a Helm chart with the properties needed to download it.
type Chart struct {
	Repository string
	ChartName  string
	Version    string
}

// ChartType represents the type of chart.
type ChartType string

const (
	ChartTypeOCI   ChartType = "oci"
	ChartTypeHTTP  ChartType = "http"
	ChartTypeLocal ChartType = "local"
)

func (c Chart) GetChartType() ChartType {
	if registry.IsOCI(c.Repository) {
		return ChartTypeOCI
	}
	if strings.HasPrefix(c.Repository, "http://") || strings.HasPrefix(c.Repository, "https://") {
		return ChartTypeHTTP
	}
	return ChartTypeLocal
}

// GetChartURL gets the artifact url of a chart version.
func (c Chart) GetChartURL(cacheDir string) (string, error) {
	// If the chart is an OCI chart, return the repository URL directly.
	if c.GetChartType() == ChartTypeOCI {
		return c.Repository, nil
	}

	chartRepo, err := repo.NewChartRepository(
		&repo.Entry{URL: c.Repository},
		getter.All(&cli.EnvSettings{}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to initialise ChartRepository '%s': %w", c.Repository, err)
	}

	chartRepo.CachePath = cacheDir

	idxContents, err := chartRepo.DownloadIndexFile()
	if err != nil {
		return "", fmt.Errorf("downloading repository '%s' index file: %w", c.Repository, err)
	}

	index, err := repo.LoadIndexFile(idxContents)
	if err != nil {
		return "", fmt.Errorf("loading repository '%s' index file: %w", c.Repository, err)
	}

	chartInfo, err := index.Get(c.ChartName, c.Version)
	if err != nil {
		return "",
			fmt.Errorf("getting chart '%s' version '%s': %w", c.Repository, c.Version, err)
	}

	if len(chartInfo.URLs) == 0 {
		return "", fmt.Errorf("chart '%s' version '%s' has no downloadable URLs", c.ChartName, c.Version)
	}

	// Chart "URLs" can be relative paths we need to resolve.
	absoluteChartURL, err := repo.ResolveReferenceURL(c.Repository, chartInfo.URLs[0])
	if err != nil {
		return "", fmt.Errorf("resolving chart URL: %w", err)
	}

	return absoluteChartURL, nil
}

// extractPath returns a normalised path to extract the chart to.
func (c Chart) extractPath() string {
	p := path.Join(c.Repository, c.ChartName)
	return strings.NewReplacer(":/", "-", ".", "-", "/", "-", ":", "-").Replace(p) + "-" + c.Version
}

// DownloadChart downloads a chart to a local directory and returns the extracted directory path.
func (c Chart) DownloadChart(cacheDir string) (string, error) {
	// Create dir if it doesn't exist
	err := os.MkdirAll(cacheDir, 0o755)
	if err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Get the full path that the chart will be extracted to.
	var chartExtractedDir string
	if registry.IsOCI(c.Repository) {
		chartExtractedDir = filepath.Join(cacheDir, c.extractPath(), filepath.Base(c.Repository))
	} else {
		chartExtractedDir = filepath.Join(cacheDir, c.extractPath(), c.ChartName)
	}

	// If extracted directory already exists, don't redownload
	if _, err := os.Stat(chartExtractedDir); err == nil {
		return chartExtractedDir, nil
	}

	// Initialise helm action config
	config := new(action.Configuration)
	if err := config.Init(nil, "", "secret", log.Printf); err != nil {
		return "", fmt.Errorf("failed to initialise helm action configuration: %w", err)
	}

	chartURL, err := c.GetChartURL(cacheDir)
	if err != nil {
		return "", fmt.Errorf("when getting chart url: %w", err)
	}

	// If the chart is an OCI chart, we need to set up registry to pull with auth
	if registry.IsOCI(c.Repository) {
		config.RegistryClient, err = registry.NewClient()
		if err != nil {
			return "", fmt.Errorf("failed to create registry client: %w", err)
		}
	}

	// Initialise pull with config
	pull := action.NewPullWithOpts(action.WithConfig(config))
	pull.Settings = cli.New()
	pull.DestDir = cacheDir
	pull.Untar = true
	pull.Version = c.Version
	pull.UntarDir = c.extractPath()

	// Download chart
	_, err = pull.Run(chartURL)
	if err != nil {
		return "", fmt.Errorf("when pulling chart: %w", err)
	}

	return chartExtractedDir, nil
}
