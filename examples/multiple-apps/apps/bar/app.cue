package kube

app: bar: {
	common: namespace: "default"

	controller: bar: {
		type: "StatefulSet"
		pod: image: "bar:latest"
		pod: ports: http: port: 8080
	}
}
