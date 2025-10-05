package kube

// Example demonstrating SOPS integration for secure secret management

// Structured secret injection from YAML file
secrets: dbCreds: _ @sops(database-creds.sops.yaml)

// Text secret injection (auto-detected for .txt files. You can specify it explicitly for yaml/json files,
// by using @sops(file.sops.yaml,type="text")
secrets: apiToken: string @sops(api-token.sops.txt)

kogen: "my-app": {
	apiVersion: "kogen.internal/v1alpha1"
	kind:       "Objects"

	spec: objects: [
		{
			apiVersion: "v1"
			kind:       "Secret"
			metadata: {
				name:      "database-credentials"
				namespace: "default"
			}
			type: "Opaque"
			stringData: {
				username: secrets.dbCreds.username
				password: secrets.dbCreds.password
				database: secrets.dbCreds.database
			}
		},
		{
			apiVersion: "v1"
			kind:       "Secret"
			metadata: {
				name:      "api-token"
				namespace: "default"
			}
			type: "Opaque"
			stringData: {
				token: secrets.apiToken
			}
		},
	]
}
