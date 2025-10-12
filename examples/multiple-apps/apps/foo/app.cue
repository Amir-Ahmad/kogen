package kube

app: foo: {
	common: namespace: "default"

	controller: foo: {
		type: "Deployment"
		container: main: {
			image: "foo:latest"
			port: http: port: 8080
		}
	}
}
