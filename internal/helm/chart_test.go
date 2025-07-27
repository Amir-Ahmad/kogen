package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPath(t *testing.T) {
	tests := map[string]struct {
		repo     string
		chart    string
		version  string
		expected string
	}{
		"http basic": {
			repo:     "https://charts.bitnami.com/bitnami",
			chart:    "nginx",
			version:  "12.0.0",
			expected: "https-charts-bitnami-com-bitnami-nginx-12.0.0",
		},
		"oci basic": {
			repo:     "oci://ghcr.io/amir-ahmad",
			chart:    "nginx",
			version:  "12.0.0",
			expected: "oci-ghcr-io-amir-ahmad-nginx-12.0.0",
		},
		"http with trailing slash": {
			repo:     "https://charts.example.com/",
			chart:    "redis",
			version:  "1.2.3",
			expected: "https-charts-example-com-redis-1.2.3",
		},
		"http with port": {
			repo:     "http://localhost:8080/charts",
			chart:    "postgres",
			version:  "2.3.4",
			expected: "http-localhost-8080-charts-postgres-2.3.4",
		},
		"nested chart name": {
			repo:     "https://repo.com/namespace",
			chart:    "myorg/mychart",
			version:  "0.1.0",
			expected: "https-repo-com-namespace-myorg-mychart-0.1.0",
		},
		"repo with dots and colons": {
			repo:     "oci://ghcr.io:443/org/chart.repo",
			chart:    "chart",
			version:  "3.3.3",
			expected: "oci-ghcr-io-443-org-chart-repo-chart-3.3.3",
		},
		"empty values": {
			repo:     "",
			chart:    "",
			version:  "",
			expected: "-",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := Chart{
				Repository: tt.repo,
				ChartName:  tt.chart,
				Version:    tt.version,
			}
			got := c.extractPath()
			assert.Equal(t, tt.expected, got)
		})
	}
}
