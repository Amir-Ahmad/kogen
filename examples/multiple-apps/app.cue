package kube

import (
	pkg_app "github.com/amir-ahmad/cue-k8s-modules/app"
)

app: [Name=string]: pkg_app.#AppConfig & {
	name: string | *Name
	// set some defaults that are specific to me - default 1 replica
	controller: [string]: X={
		if X.type == "Deployment" || X.type == "StatefulSet" {
			spec: replicas: int | *1
		}
	}
}

kogen: [string]: {
	apiVersion: string
	kind:       string
	spec: {...}
}

for k, v in app {
	kogen: "\(k)": {
		apiVersion: "kogen.internal/v1alpha1"
		kind:       "Objects"
		spec: objects: (pkg_app.#App & {config: v}).out
	}
}
