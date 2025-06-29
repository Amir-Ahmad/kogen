package cmd

import (
	"fmt"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type BuildCmd struct {
	Path string `short:"p" help:"Path to read manifests from."`
}

func (b *BuildCmd) Run() error {
	manifests, err := parseYamlManifests(os.Stdin)
	if err != nil {
		return err
	}
	for _, manifest := range manifests {
		fmt.Println("parsed a manifest")
		fmt.Printf("%+v\n", manifest)
		fmt.Printf("kind: %s\n", manifest.GetKind())
		fmt.Printf("apiVersion: %s\n", manifest.GetAPIVersion())
	}
	return nil
}

func parseYamlManifests(r io.Reader) ([]unstructured.Unstructured, error) {
	decoder := yaml.NewYAMLToJSONDecoder(r)
	var manifests []unstructured.Unstructured
	for {
		var manifest unstructured.Unstructured
		if err := decoder.Decode(&manifest); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		manifests = append(manifests, manifest)
	}
	return manifests, nil
}
