package sops

import (
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFilePath(t *testing.T) {
	tests := map[string]struct {
		filename    string
		instanceDir string
		moduleRoot  string
		expected    string
	}{
		"relative path uses instance dir": {
			filename:    "secrets.yaml",
			instanceDir: "/path/to/instance",
			moduleRoot:  "/path/to/module",
			expected:    "/path/to/instance/secrets.yaml",
		},
		"module prefix uses module root": {
			filename:    "module://secrets.yaml",
			instanceDir: "/path/to/instance",
			moduleRoot:  "/path/to/module",
			expected:    "/path/to/module/secrets.yaml",
		},
		"module prefix with subdirectory": {
			filename:    "module://config/secrets.yaml",
			instanceDir: "/path/to/instance",
			moduleRoot:  "/path/to/module",
			expected:    "/path/to/module/config/secrets.yaml",
		},
		"absolute path": {
			filename:    "/absolute/path/secrets.yaml",
			instanceDir: "/path/to/instance",
			moduleRoot:  "/path/to/module",
			expected:    "/path/to/instance/absolute/path/secrets.yaml",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := parseFilePath(tc.filename, tc.instanceDir, tc.moduleRoot)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFindSopsAttribute(t *testing.T) {
	ctx := cuecontext.New()

	tests := map[string]struct {
		cueInput      string
		expectedCfg   sopsConfig
		expectedFound bool
	}{
		"no sops attribute": {
			cueInput:      `field: "value"`,
			expectedCfg:   sopsConfig{},
			expectedFound: false,
		},
		"simple sops attribute": {
			cueInput: `field: _ @sops(secrets.yaml)`,
			expectedCfg: sopsConfig{
				filename: "secrets.yaml",
				textMode: false,
			},
			expectedFound: true,
		},
		"sops attribute with text mode": {
			cueInput: `field: _ @sops(secrets.yaml,type=text)`,
			expectedCfg: sopsConfig{
				filename: "secrets.yaml",
				textMode: true,
			},
			expectedFound: true,
		},
		"sops attribute with module prefix": {
			cueInput: `field: _ @sops("module://secrets.yaml")`,
			expectedCfg: sopsConfig{
				filename: "module://secrets.yaml",
				textMode: false,
			},
			expectedFound: true,
		},
		"sops attribute with text mode and module prefix": {
			cueInput: `field: _ @sops("module://creds.txt",type=text)`,
			expectedCfg: sopsConfig{
				filename: "module://creds.txt",
				textMode: true,
			},
			expectedFound: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			v := ctx.CompileString(tc.cueInput)
			require.NoError(t, v.Err())

			// Get the field value
			iter, err := v.Fields()
			require.NoError(t, err)
			require.True(t, iter.Next())
			fieldValue := iter.Value()

			cfg, found := findSopsAttribute(fieldValue)
			assert.Equal(t, tc.expectedFound, found)
			if tc.expectedFound {
				assert.Equal(t, tc.expectedCfg.filename, cfg.filename)
				assert.Equal(t, tc.expectedCfg.textMode, cfg.textMode)
			}
		})
	}
}
