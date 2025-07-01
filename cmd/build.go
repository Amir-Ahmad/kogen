package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/amir-ahmad/kogen/internal/build"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

type BuildCmd struct {
	Chdir   string   `short:"c" help:"Change directory before running"`
	Path    string   `short:"p" help:"Path to read manifests from. Specify '-' to read yaml from stdin." required:""`
	Tag     []string `short:"t" help:"Tags to pass to Cue"`
	Package string   `help:"Package to load in Cue"`
}

func (b *BuildCmd) Run() error {
	if b.Chdir != "" {
		if err := os.Chdir(b.Chdir); err != nil {
			return fmt.Errorf("failed to change directory: %w", err)
		}
	}

	manifests, err := b.readManifests(b.Path)
	if err != nil {
		return err
	}

	for _, manifest := range manifests {
		fmt.Println("parsed a manifest")
		fmt.Printf("%+v\n", manifest)
		fmt.Printf("kind: %s\n", manifest.GetKind())
		fmt.Printf("apiVersion: %s\n", manifest.GetAPIVersion())
		fmt.Printf("gvk: %s\n", manifest.GroupVersionKind())
	}
	return build.Generate(os.Stdout, manifests)
}

func (b *BuildCmd) readManifests(path string) ([]unstructured.Unstructured, error) {
	if path == "-" {
		return readYamlManifests(os.Stdin)
	}

	if strings.HasSuffix(strings.ToLower(path), ".yaml") || strings.HasSuffix(strings.ToLower(path), ".yml") {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		return readYamlManifests(f)
	}

	return b.readCueManifests(path)
}

func readYamlManifests(r io.Reader) ([]unstructured.Unstructured, error) {
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

func (b *BuildCmd) readCueManifests(path string) ([]unstructured.Unstructured, error) {
	ctx := cuecontext.New()
	cfg := load.Config{Tags: b.Tag}
	if b.Package != "" {
		cfg.Package = b.Package
	}

	manifests := []unstructured.Unstructured{}

	insts := load.Instances([]string{path}, &cfg)
	for _, inst := range insts {
		if inst.Err != nil {
			return nil, fmt.Errorf("error when loading cue instance: %w", inst.Err)
		}

		v := ctx.BuildInstance(inst)
		if v.Err() != nil {
			return nil, fmt.Errorf("failed to build cue instance: %w", v.Err())
		}

		cogValue := v.LookupPath(cue.ParsePath("cog"))
		if cogValue.Exists() {
			if cogValue.Err() != nil {
				return nil, fmt.Errorf("failed to parse cog: %w", cogValue.Err())
			}

			var manifest unstructured.Unstructured
			if err := cogValue.Decode(&manifest); err != nil {
				return nil, fmt.Errorf("failed to decode cog: %w", err)
			}

			manifests = append(manifests, manifest)
		}

		cogsValue := v.LookupPath(cue.ParsePath("cogList"))
		if cogsValue.Exists() {
			if cogsValue.Err() != nil {
				return nil, fmt.Errorf("failed to parse cogList: %w", cogsValue.Err())
			}

			var cogs []unstructured.Unstructured
			if err := cogsValue.Decode(&cogs); err != nil {
				return nil, fmt.Errorf("failed to decode cogList: %w", err)
			}

			manifests = append(manifests, cogs...)
		}

		fmt.Printf("loaded instance: %v\n", inst)
		fmt.Printf("val: %+v\n", v)
	}
	return manifests, nil
}
