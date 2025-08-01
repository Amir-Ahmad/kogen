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
}
