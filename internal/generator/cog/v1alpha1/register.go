package v1alpha1

import (
	"github.com/amir-ahmad/kogen/api/v1alpha1"
	"github.com/amir-ahmad/kogen/internal/registry"
)

func init() {
	registry.Register(v1alpha1.CogGVK, NewGenerator)
}
