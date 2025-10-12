package kube

app: bar: {
	common: namespace: "default"

	controller: bar: {
		type: "StatefulSet"
		container: main: {
			image: "bar:latest"
			port: http: port: 8080
		}
	}
}
