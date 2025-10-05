package v1alpha1

import (
	"bytes"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/amir-ahmad/kogen/internal/generator"
	"github.com/stretchr/testify/require"
)

func TestOutput_EncodeError(t *testing.T) {
	// Create an incomplete/error cue value
	ctx := cuecontext.New()
	val := ctx.CompileString("_")

	obj := Object{
		kind:  "ConfigMap",
		value: val,
	}

	var buf bytes.Buffer
	err := obj.Output(&buf)
	require.ErrorContains(t, err, "failed to encode object to yaml")
}

func TestGenerate_ExitEarly(t *testing.T) {
	ctx := cuecontext.New()
	input := `
		objects: [
			{
				apiVersion: "v1"
				kind: "ConfigMap"
				metadata: name: "test1"
			},
			{
				apiVersion: "v1"
				kind: "Secret"
				metadata: name: "test2"
			}
		]
	`
	val := ctx.CompileString(input)
	require.NoError(t, val.Err(), "failed to compile cue")

	gen := &Generator{
		objects: val.LookupPath(cue.MakePath(cue.Str("objects"))),
	}

	iter, err := gen.Generate(generator.Options{})
	require.NoError(t, err, "failed to generate objects")

	// Iterate only once to trigger early termination
	count := 0
	for _, err := range iter {
		require.NoError(t, err)
		count++
		if count == 1 {
			break
		}
	}

	require.Equal(t, 1, count)
}
