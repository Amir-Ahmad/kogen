package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var GroupVersion = schema.GroupVersion{
	Group:   "kogen.internal",
	Version: "v1alpha1",
}

var CogGVK = GroupVersion.WithKind("Cog")
var ObjectsGVK = GroupVersion.WithKind("Objects")
