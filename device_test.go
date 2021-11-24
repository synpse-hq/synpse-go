package synpse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListDevices(t *testing.T) {
	client := getTestingProjectClient(t)

	devicesResp, err := client.ListDevices(context.Background(), &ListDevicesRequest{})
	require.NoError(t, err)

	assert.True(t, len(devicesResp.Devices) > 0)
}

func TestDevices(t *testing.T) {
	client := getTestingProjectClient(t)
	ctx := context.Background()

	var (
		// Will be used for device registration token
		drtName = testDeviceRegistrationPrefix + ksuid.New().String()
		// Will be set once the device is registered
		deviceID = ""
		// Test device
		testingDevice = Device{}
	)

	drt, err := client.CreateRegistrationToken(ctx, DeviceRegistrationToken{
		Name:                 drtName,
		MaxRegistrations:     toInt(1),
		Labels:               map[string]string{"foo": "bar"},
		EnvironmentVariables: map[string]string{"FOO": "BAR"},
	})
	require.NoError(t, err, "failed to create device registration token")

	t.Cleanup(func() {
		_ = client.DeleteRegistrationToken(ctx, drt.ID)
	})

	t.Run("RegisterDevice", func(t *testing.T) {
		registered := registerDevice(t, client, &registerDeviceRequest{
			DeviceRegistrationTokenID: drt.ID,
			DeviceInfo: DeviceInfo{
				Hostname: "test-hostname",
				OSRelease: OSRelease{
					Name: "test-os-release",
				},
			},
		})

		// Setting up registered device
		deviceID = registered.DeviceID

		assert.NotEmpty(t, registered.DeviceID)
	})

	t.Run("CreatedDeviceFoundInList", func(t *testing.T) {
		devicesResp, err := client.ListDevices(context.Background(), &ListDevicesRequest{})
		require.NoError(t, err)

		assert.True(t, len(devicesResp.Devices) > 0)

		deviceFound := false
		for _, device := range devicesResp.Devices {
			t.Logf("device: '%s'", device.Name)
			if device.ID == deviceID {
				deviceFound = true
				testingDevice = *device
				break
			}
		}

		require.True(t, deviceFound, "testing device not found, have you registered it?")
	})

	t.Run("AddLabel", func(t *testing.T) {
		if testingDevice.Labels == nil {
			testingDevice.Labels = make(map[string]string)
		}
		testingDevice.Labels["new"] = "label"
		_, err := client.UpdateDevice(context.Background(), testingDevice)
		require.NoError(t, err, "failed to update device")

		// Getting the device
		stored, err := client.GetDevice(context.Background(), testingDevice.Name)
		require.NoError(t, err, "failed to get device")
		assert.Equal(t, "label", stored.Labels["new"])
	})

	t.Run("DeleteDevice", func(t *testing.T) {

		err := client.DeleteDevice(context.Background(), testingDevice.Name)
		require.NoError(t, err, "failed to update device")

		// Getting the device
		_, err = client.GetDevice(context.Background(), testingDevice.Name)
		require.Error(t, err, "device should have been deleted")
	})
}

func registerDevice(t *testing.T, apiClient *API, deviceReq *registerDeviceRequest) *registerDeviceResponse {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(deviceReq)
	require.NoError(t, err)

	fmt.Println(getURL(apiClient.BaseURL, projectsURL, apiClient.ProjectID, devicesURL))

	req, err := http.NewRequest(http.MethodPost, getURL(apiClient.BaseURL, projectsURL, apiClient.ProjectID, devicesURL, "register"), &buf)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var registerResp registerDeviceResponse
	err = json.NewDecoder(resp.Body).Decode(&registerResp)
	require.NoError(t, err)

	return &registerResp
}

type registerDeviceRequest struct {
	DeviceRegistrationTokenID string     `json:"deviceRegistrationTokenId"`
	DeviceInfo                DeviceInfo `json:"deviceInfo"`
}

type registerDeviceResponse struct {
	DeviceID             string `json:"deviceId"`
	DeviceAccessKeyValue string `json:"deviceAccessKeyValue"`
}
