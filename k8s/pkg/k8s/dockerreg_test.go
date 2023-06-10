package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDockerRegSecret(t *testing.T) {
	secret := CreateDockerRegistrySecret("username", "password", "server")
	secretSerialized, err := secret.Serialize()
	assert.Nil(t, err)

	assert.Equal(t, secretSerialized, "{\"auths\":{\"server\":{\"username\":\"username\",\"password\":\"password\",\"auth\":\"dXNlcm5hbWU6cGFzc3dvcmQ=\"}}}")
}
