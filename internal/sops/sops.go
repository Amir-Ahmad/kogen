package sops

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"cuelang.org/go/cue"
	"github.com/getsops/sops/v3/cmd/sops/formats"
	"github.com/getsops/sops/v3/decrypt"
	"sigs.k8s.io/yaml"
)

const (
	SopsAttribute = "sops"
)

func Inject(v cue.Value, sopsPath cue.Path, instanceDir string) (cue.Value, error) {
	secretsValue := v.LookupPath(sopsPath)
	if !secretsValue.Exists() {
		return v, nil
	}

	filledSecrets, changed, err := fillSecrets(secretsValue, instanceDir)
	if err != nil {
		return cue.Value{}, fmt.Errorf("failed to process @sops attributes: %w", err)
	}

	if changed {
		v = v.FillPath(sopsPath, filledSecrets)
		if err := v.Err(); err != nil {
			return cue.Value{}, fmt.Errorf("failed to inject decrypted secrets into cue value: %w", err)
		}
	}

	return v, nil
}

func fillSecrets(v cue.Value, instanceDir string) (cue.Value, bool, error) {
	if sopsFile := findSopsAttribute(v); sopsFile != "" {
		result, err := replaceBySopsDecryption(v, instanceDir, sopsFile)
		return result, true, err
	}
	return processFieldsRecursively(v, instanceDir)
}

func findSopsAttribute(v cue.Value) string {
	for _, attr := range v.Attributes(cue.FieldAttr) {
		if attr.Name() == SopsAttribute {
			return attr.Contents()
		}
	}
	return ""
}

func replaceBySopsDecryption(v cue.Value, instanceDir string, sopsFile string) (cue.Value, error) {
	filePath := filepath.Join(instanceDir, sopsFile)

	content, err := getDecryptedContent(filePath)
	if err != nil {
		return cue.Value{}, err
	}

	contentValue := v.Context().Encode(content)
	if err := contentValue.Err(); err != nil {
		return cue.Value{}, fmt.Errorf("failed to encode decrypted content from %q as cue value: %w", filePath, err)
	}
	return contentValue, nil
}

func processFieldsRecursively(v cue.Value, instanceDir string) (cue.Value, bool, error) {
	fields, err := v.Fields(cue.Optional(true), cue.Hidden(true), cue.Definitions(true))
	if err != nil {
		return cue.Value{}, false, fmt.Errorf("failed to iterate cue fields: %w", err)
	}

	result := v
	changed := false
	for fields.Next() {
		selector := fields.Selector()
		fieldValue := fields.Value()

		processedValue, fieldChanged, err := fillSecrets(fieldValue, instanceDir)
		if err != nil {
			return cue.Value{}, false, fmt.Errorf("in field %s: %w", selector, err)
		}

		if fieldChanged {
			result = result.FillPath(cue.MakePath(selector), processedValue)
			if err := result.Err(); err != nil {
				return cue.Value{}, false, fmt.Errorf("failed to inject secrets into field %s: %w", selector, err)
			}
			changed = true
		}
	}

	return result, changed, nil
}

func getDecryptedContent(file string) (any, error) {
	var format string
	if formats.IsYAMLFile(file) {
		format = "yaml"
	} else if formats.IsJSONFile(file) {
		format = "json"
	} else {
		return nil, fmt.Errorf("sops file %q must be .yaml, .yml, or .json", file)
	}

	decryptedBytes, err := decrypt.File(file, format)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt sops file %q: %w", file, err)
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
