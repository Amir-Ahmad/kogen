package kube

kogen: hello: {
	apiVersion: "kogen.internal/v1alpha1"
	kind:       "Cog"

	spec: helm: [{
		releaseName: "hello-world"
		repository:  "https://helm.github.io/examples"
		chartName:   "hello-world"
		version:     "0.1.0"
		values: replicaCount: 2
	}]

	// Use kustomize to add labels to all helm resources
	// All kustomize properties are supported under the kustomize field including patches.
	spec: kustomize: labels: [{
		pairs: team: "kogen-test"
	}]
}
