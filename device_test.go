package synpse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

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
		// Labels will be applied to the DRT
		drtLabels = map[string]string{
			"foo": "bar",
		}
		// Will be set once the device is registered
		deviceID = ""
		// Test device
		testingDevice = Device{}
	)

	drt, err := client.CreateRegistrationToken(ctx, DeviceRegistrationToken{
		Name:                 drtName,
		MaxRegistrations:     toInt(1),
		Labels:               drtLabels,
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
		devicesResp, err := client.ListDevices(context.Background(), &ListDevicesRequest{
			Labels: drtLabels,
		})
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

	t.Run("DeviceFilteringByLabels", func(t *testing.T) {
		devicesResp, err := client.ListDevices(context.Background(), &ListDevicesRequest{
			Labels: map[string]string{
				"shouldnotbethere": "val",
			},
		})
		require.NoError(t, err)

		assert.True(t, len(devicesResp.Devices) == 0, "expected to not find any devices")
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
		require.NoError(t, err, "failed to delete device")

		// Getting the device
		_, err = client.GetDevice(context.Background(), testingDevice.Name)
		require.Error(t, err, "device should have been deleted")
	})
}

func TestDeviceFilteringAndPagination(t *testing.T) {
	// Create a lot of devices
	client := getTestingProjectClient(t)
	ctx := context.Background()

	devicesPerGroup := 10

	var (
		groupOneDeviceIDs []string
		groupTwoDeviceIDs []string
	)

	// Will be used for device registration token
	drtOneName := testDeviceRegistrationPrefix + ksuid.New().String()
	drtTwoName := testDeviceRegistrationPrefix + ksuid.New().String()

	sharedLabelValue := ksuid.New().String()
	groupOneLabelValue := ksuid.New().String()
	groupTwoLabelValue := ksuid.New().String()

	drtGroupOne, err := client.CreateRegistrationToken(ctx, DeviceRegistrationToken{
		Name:             drtOneName,
		MaxRegistrations: toInt(devicesPerGroup),
		Labels: map[string]string{
			"shared": sharedLabelValue,
			"group":  groupOneLabelValue,
		},
	})
	require.NoError(t, err, "failed to create device registration token")
	t.Cleanup(func() {
		_ = client.DeleteRegistrationToken(ctx, drtGroupOne.ID)
	})

	drtGroupTwo, err := client.CreateRegistrationToken(ctx, DeviceRegistrationToken{
		Name:             drtTwoName,
		MaxRegistrations: toInt(devicesPerGroup),
		Labels: map[string]string{
			"shared": sharedLabelValue,
			"group":  groupTwoLabelValue,
		},
	})
	require.NoError(t, err, "failed to create device registration token")
	t.Cleanup(func() {
		_ = client.DeleteRegistrationToken(ctx, drtGroupTwo.ID)
	})

	// Registering devices
	for i := 0; i < devicesPerGroup; i++ {
		// Group one
		registeredGroupOne := registerDevice(t, client, &registerDeviceRequest{
			DeviceRegistrationTokenID: drtGroupOne.ID,
			DeviceInfo: DeviceInfo{
				Hostname: "test-hostname",
				OSRelease: OSRelease{
					Name: "test-os-release",
				},
			},
		})
		time.Sleep(500 * time.Millisecond)
		groupOneDeviceIDs = append(groupOneDeviceIDs, registeredGroupOne.DeviceID)

		t.Cleanup(func() {
			err := client.DeleteDevice(context.Background(), registeredGroupOne.DeviceID)
			require.NoError(t, err, "failed to delete device")
		})
		// Group two
		registeredGroupTwo := registerDevice(t, client, &registerDeviceRequest{
			DeviceRegistrationTokenID: drtGroupTwo.ID,
			DeviceInfo: DeviceInfo{
				Hostname: "test-hostname",
				OSRelease: OSRelease{
					Name: "test-os-release",
				},
			},
		})
		time.Sleep(500 * time.Millisecond)
		groupTwoDeviceIDs = append(groupTwoDeviceIDs, registeredGroupTwo.DeviceID)

		t.Cleanup(func() {
			err := client.DeleteDevice(context.Background(), registeredGroupTwo.DeviceID)
			require.NoError(t, err, "failed to delete device")
		})
	}

	t.Run("ListAndValidateAll", func(t *testing.T) {
		// List all devices
		devicesResp, err := client.ListDevices(context.Background(), &ListDevicesRequest{
			Labels: map[string]string{
				"shared": sharedLabelValue,
			},
		})
		require.NoError(t, err)
		require.Len(t, devicesResp.Devices, devicesPerGroup*2, "unexpected number of devices")

		// Ensure all are in
		matchDevicesAndDeviceIDs(t, append(groupOneDeviceIDs, groupTwoDeviceIDs...), devicesResp.Devices)
	})

	t.Run("ListAndValidateGroupOne", func(t *testing.T) {
		devicesResp, err := client.ListDevices(context.Background(), &ListDevicesRequest{
			Labels: map[string]string{
				"group": groupOneLabelValue,
			},
		})
		require.NoError(t, err)
		require.Len(t, devicesResp.Devices, devicesPerGroup, "unexpected number of devices")

		matchDevicesAndDeviceIDs(t, groupOneDeviceIDs, devicesResp.Devices)
	})

	t.Run("ListAndValidateGroupTwo", func(t *testing.T) {
		devicesResp, err := client.ListDevices(context.Background(), &ListDevicesRequest{
			Labels: map[string]string{
				"group": groupTwoLabelValue,
			},
		})
		require.NoError(t, err)
		require.Len(t, devicesResp.Devices, devicesPerGroup, "unexpected number of devices")

		matchDevicesAndDeviceIDs(t, groupTwoDeviceIDs, devicesResp.Devices)
	})
}

func matchDevicesAndDeviceIDs(t *testing.T, deviceIDs []string, devices []*Device) {
	assert.Len(t, deviceIDs, len(devices))
	for _, device := range devices {
		assert.Contains(t, deviceIDs, device.ID)
	}
}

func registerDevice(t *testing.T, apiClient *API, deviceReq *registerDeviceRequest) *registerDeviceResponse {
	t.Helper()

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(deviceReq)
	require.NoError(t, err)

	fmt.Println(getURL(apiClient.BaseURL, projectsURL, apiClient.ProjectID, devicesURL))

	req, err := http.NewRequest(http.MethodPost, getURL(apiClient.BaseURL, projectsURL, apiClient.ProjectID, devicesURL, "register"), &buf)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bts, _ := ioutil.ReadAll(resp.Body)
		require.Equal(t, http.StatusOK, resp.StatusCode, string(bts))
	}

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
