package k8s

import (
	"encoding/base64"
	"encoding/json"
)

type DockerRegistrySecret struct {
	DockerCredentials map[string]DockerCredential `json:"auths"`
}

type DockerCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth"`
}

func CreateDockerRegistrySecret(username string, password string, regURL string) *DockerRegistrySecret {
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

	credentials := make(map[string]DockerCredential)
	credential := DockerCredential{Username: username, Password: password, Auth: auth}
	credentials[regURL] = credential
	return &DockerRegistrySecret{DockerCredentials: credentials}
}

func (secret *DockerRegistrySecret) Serialize() (string, error) {
	jsonBytes, err := json.Marshal(secret)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
