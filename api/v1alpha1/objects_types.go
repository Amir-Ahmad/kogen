package v1alpha1

import (
	"cuelang.org/go/cue"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Objects struct {
	metav1.TypeMeta `json:",inline"`
	Spec            ObjectsSpec `json:"spec"`
}

type ObjectsSpec struct {
	Objects cue.Value `json:"objects"`
}
