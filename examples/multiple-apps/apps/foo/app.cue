package kube

app: foo: {
	common: namespace: "default"

	controller: foo: {
		type: "Deployment"
		pod: image: "foo:latest"
		pod: ports: http: port: 8080
	}
}
