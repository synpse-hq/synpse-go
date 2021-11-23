package synpse

import (
	"context"
	"testing"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO
const testDeviceRegistrationPrefix = "sdk-drt-test-"

func TestListDeviceRegistrationTokens(t *testing.T) {
	client := getTestingProjectClient(t)

	tokens, err := client.ListDeviceRegistrationTokens(context.Background(), &ListDeviceRegistrationTokensRequest{})
	require.NoError(t, err)

	assert.True(t, len(tokens) > 0)
}

func TestDeviceRegistrationTokens(t *testing.T) {
	client := getTestingProjectClient(t)
	ctx := context.Background()

	drtName := testDeviceRegistrationPrefix + ksuid.New().String()

	drt, err := client.CreateRegistrationToken(ctx, DeviceRegistrationToken{
		Name:                 drtName,
		MaxRegistrations:     toInt(1),
		Labels:               map[string]string{"foo": "bar"},
		EnvironmentVariables: map[string]string{"FOO": "BAR"},
	})
	require.NoError(t, err, "failed to create device registration token")

	assert.Equal(t, drtName, drt.Name)

	t.Run("FindCreatedDeviceRegistrationToken", func(t *testing.T) {
		tokens, err := client.ListDeviceRegistrationTokens(context.Background(), &ListDeviceRegistrationTokensRequest{})
		require.NoError(t, err)

		tokenFound := false
		for _, token := range tokens {
			if token.Name == drtName {
				tokenFound = true
				assert.Equal(t, drt.Name, token.Name)
				assert.Equal(t, drt.MaxRegistrations, token.MaxRegistrations)
				assert.Equal(t, drt.Labels, token.Labels)
				assert.Equal(t, drt.EnvironmentVariables, token.EnvironmentVariables)
			}
		}
		assert.True(t, tokenFound, "created token not found")
	})

	t.Run("UpdateDeviceRegistrationToken", func(t *testing.T) {
		drt, err := client.UpdateRegistrationToken(ctx, DeviceRegistrationToken{
			ID:                   drt.ID,
			Name:                 drtName,
			MaxRegistrations:     toInt(2),
			Labels:               map[string]string{"foo2": "bar2"},
			EnvironmentVariables: map[string]string{"FOO2": "BAR2"},
		})
		require.NoError(t, err, "failed to update device registration token")

		assert.Equal(t, drtName, drt.Name)
		assert.Equal(t, toInt(2), drt.MaxRegistrations)
		assert.Equal(t, map[string]string{"foo2": "bar2"}, drt.Labels)
		assert.Equal(t, map[string]string{"FOO2": "BAR2"}, drt.EnvironmentVariables)
	})

	t.Run("DeleteDeviceRegistrationToken", func(t *testing.T) {
		err := client.DeleteRegistrationToken(ctx, drt.ID)
		require.NoError(t, err, "failed to delete device registration token")
	})
}

func toInt(v int) *int {
	return &v
}
