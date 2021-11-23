package synpse

import (
	"context"
	"testing"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecretPrefix = "sdk-test-"

func TestListSecrets(t *testing.T) {
	client := getTestingProjectClient(t)

	_, err := client.ListSecrets(context.Background(), &ListSecretsRequest{Namespace: sdkTestNamespace})
	require.NoError(t, err, "failed to list secrets")
}

func TestEnvironmentSecret(t *testing.T) {
	client := getTestingProjectClient(t)

	secretName := testSecretPrefix + ksuid.New().String()

	environmentSecret, err := client.CreateSecret(context.Background(), sdkTestNamespace, Secret{
		Name: secretName,
		Type: SecretTypeEnvironment,
		Data: "some-data",
	})
	require.NoError(t, err, "failed to create secret")

	t.Logf("created secret %s (%s)", environmentSecret.Name, environmentSecret.ID)

	t.Run("FindCreatedSecret", func(t *testing.T) {
		secrets, err := client.ListSecrets(context.Background(), &ListSecretsRequest{Namespace: sdkTestNamespace})
		require.NoError(t, err)

		secretFound := false
		for _, secret := range secrets {
			if secret.Name == secretName {
				secretFound = true
				break
			}
		}
		require.True(t, secretFound, "created secret not found")
	})

	t.Run("CheckSecretDetails", func(t *testing.T) {
		stored, err := client.GetSecret(context.Background(), sdkTestNamespace, secretName)
		require.NoError(t, err)

		require.Equal(t, environmentSecret.ID, stored.ID)
		assert.Equal(t, "some-data", stored.Data, "secret data doesn't match")
	})

	t.Run("DeleteSecret", func(t *testing.T) {
		err := client.DeleteSecret(context.Background(), sdkTestNamespace, secretName)
		require.NoError(t, err)

		_, err = client.GetSecret(context.Background(), sdkTestNamespace, secretName)
		require.Error(t, err, "expected to get an error")
	})
}
