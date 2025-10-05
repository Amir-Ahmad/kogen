package sops

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"github.com/getsops/sops/v3/cmd/sops/formats"
	"github.com/getsops/sops/v3/decrypt"
	"sigs.k8s.io/yaml"
)

const (
	SopsAttribute = "sops"
	ModulePrefix  = "module://"
)

func Inject(v cue.Value, sopsPath cue.Path, instanceDir, moduleRoot string) (cue.Value, error) {
	secretsValue := v.LookupPath(sopsPath)
	if !secretsValue.Exists() {
		return v, nil
	}

	filledSecrets, changed, err := fillSecrets(secretsValue, instanceDir, moduleRoot)
	if err != nil {
		return cue.Value{}, fmt.Errorf("failed to process @sops attributes: %w", err)
	}

	if changed {
		v = v.FillPath(sopsPath, filledSecrets)
		if err := v.Err(); err != nil {
			return cue.Value{}, fmt.Errorf(
				"failed to inject decrypted secrets into cue value: %w",
				err,
			)
		}
	}

	return v, nil
}

func fillSecrets(v cue.Value, instanceDir, moduleRoot string) (cue.Value, bool, error) {
	if config, found := findSopsAttribute(v); found {
		result, err := replaceBySopsDecryption(v, instanceDir, moduleRoot, config)
		return result, true, err
	}
	return processFieldsRecursively(v, instanceDir, moduleRoot)
}

type sopsConfig struct {
	filename string
	textMode bool
}

func findSopsAttribute(v cue.Value) (sopsConfig, bool) {
	for _, attr := range v.Attributes(cue.FieldAttr) {
		if attr.Name() == SopsAttribute {
			// First positional argument is the filename
			filename, err := attr.String(0)
			if err != nil {
				// Fall back to raw contents if parsing fails
				return sopsConfig{filename: attr.Contents()}, true
			}

			// Check for type=text parameter
			typeVal, found, _ := attr.Lookup(1, "type")
			textMode := found && typeVal == "text"

			return sopsConfig{filename: filename, textMode: textMode}, true
		}
	}
	return sopsConfig{}, false
}

// parseFilePath resolves a file path, handling the module:// prefix.
// If the path starts with module://, it's resolved relative to moduleRoot.
// Otherwise, it's resolved relative to instanceDir.
func parseFilePath(filename, instanceDir, moduleRoot string) string {
	if strings.HasPrefix(filename, ModulePrefix) {
		// Strip the module:// prefix and resolve relative to module root
		relativePath := strings.TrimPrefix(filename, ModulePrefix)
		return filepath.Join(moduleRoot, relativePath)
	}
	// Default: resolve relative to instance directory
	return filepath.Join(instanceDir, filename)
}

func replaceBySopsDecryption(
	v cue.Value,
	instanceDir, moduleRoot string,
	config sopsConfig,
) (cue.Value, error) {
	filePath := parseFilePath(config.filename, instanceDir, moduleRoot)

	content, err := getDecryptedContent(filePath, config.textMode)
	if err != nil {
		return cue.Value{}, err
	}

	contentValue := v.Context().Encode(content)
	if err := contentValue.Err(); err != nil {
		return cue.Value{}, fmt.Errorf(
			"failed to encode decrypted content from %q as cue value: %w",
			filePath,
			err,
		)
	}
	return contentValue, nil
}

func processFieldsRecursively(
	v cue.Value,
	instanceDir, moduleRoot string,
) (cue.Value, bool, error) {
	fields, err := v.Fields(cue.Optional(true), cue.Hidden(true), cue.Definitions(true))
	if err != nil {
		return cue.Value{}, false, fmt.Errorf("failed to iterate cue fields: %w", err)
	}

	result := v
	changed := false
	for fields.Next() {
		selector := fields.Selector()
		fieldValue := fields.Value()

		processedValue, fieldChanged, err := fillSecrets(fieldValue, instanceDir, moduleRoot)
		if err != nil {
			return cue.Value{}, false, fmt.Errorf("in field %s: %w", selector, err)
		}

		if fieldChanged {
			result = result.FillPath(cue.MakePath(selector), processedValue)
			if err := result.Err(); err != nil {
				return cue.Value{}, false, fmt.Errorf(
					"failed to inject secrets into field %s: %w",
					selector,
					err,
				)
			}
			changed = true
		}
	}

	return result, changed, nil
}

func getDecryptedContent(file string, textMode bool) (any, error) {
	var format string
	isYAML := formats.IsYAMLFile(file)
	isJSON := formats.IsJSONFile(file)

	// Determine if we should use text mode
	useTextMode := textMode || (!isYAML && !isJSON)

	if isYAML {
		format = "yaml"
	} else if isJSON {
		format = "json"
	} else {
		// For non-YAML/JSON files, use binary format
		format = "binary"
	}

	decryptedBytes, err := decrypt.File(file, format)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt sops file %q: %w", file, err)
	}

	// If text mode is requested or file is not structured, return as string
	if useTextMode {
		return string(decryptedBytes), nil
	}

	var result any

	switch format {
	case "yaml":
		err = yaml.Unmarshal(decryptedBytes, &result)
		if err != nil {
			return nil, fmt.Errorf("failed to parse decrypted yaml from %q: %w", file, err)
		}
	case "json":
		err = json.Unmarshal(decryptedBytes, &result)
		if err != nil {
			return nil, fmt.Errorf("failed to parse decrypted json from %q: %w", file, err)
		}
	}

	return result, nil
}
