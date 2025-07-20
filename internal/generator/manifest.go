package generator

import (
	"cuelang.org/go/cue"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Manifest represents a generator manifest.
type Manifest struct {
	metav1.TypeMeta `json:",inline"`
	Spec            cue.Value
}
