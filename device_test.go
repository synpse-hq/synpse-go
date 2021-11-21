package synpse

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// const testDeviceName = "testing-device"
const testDeviceName = "silly-haibt"

func TestListDevices(t *testing.T) {
	client := getTestingProjectClient(t)

	devices, err := client.ListDevices(context.Background(), []string{})
	require.NoError(t, err)

	assert.True(t, len(devices) > 0)
}

func TestDevices(t *testing.T) {
	client := getTestingProjectClient(t)

	devices, err := client.ListDevices(context.Background(), []string{})
	require.NoError(t, err)

	assert.True(t, len(devices) > 0)

	deviceFound := false
	var testingDevice *Device
	for _, device := range devices {
		t.Logf("device: '%s'", device.Name)
		if device.Name == testDeviceName {
			testingDevice = device
			deviceFound = true
			break
		}
	}

	require.True(t, deviceFound, "testing device not found, have you registered it?")

	t.Run("AddLabel", func(t *testing.T) {
		if testingDevice.Labels == nil {
			testingDevice.Labels = make(map[string]string)
		}
		testingDevice.Labels["foo"] = "bar"
		_, err := client.UpdateDevice(context.Background(), *testingDevice)
		require.NoError(t, err, "failed to update device")

		// Getting the device
		stored, err := client.GetDevice(context.Background(), testDeviceName)
		require.NoError(t, err, "failed to get device")
		assert.Equal(t, "bar", stored.Labels["foo"])
	})
}
