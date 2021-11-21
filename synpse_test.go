package synpse

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	EnvSynpseAPIEndpointURL = "SYNPSE_API_ENDPOINT_URL"

	EnvSynpsePersonalAccessKey = "SYNPSE_PERSONAL_ACCESS_KEY"
	EnvSynpseProjectAccessKey  = "SYNPSE_PROJECT_ACCESS_KEY"

	EnvSynpseSDKTestProjectName = "SYNPSE_SDK_TEST_PROJECT_NAME"
	EnvSynpseSDKTestProjectID   = "SYNPSE_SDK_TEST_PROJECT_ID"

	EnvSynpseSDKTestApplicationName = "SYNPSE_SDK_TEST_APPLICATION_NAME"
)

// Testing data
var (
	sdkTestProjectName     string
	sdkTestProjectID       string
	sdkTestApplicationName string
	sdkTestNamespace       = "default"
)

func init() {
	sdkTestProjectName = os.Getenv(EnvSynpseSDKTestProjectName)
	sdkTestProjectID = os.Getenv(EnvSynpseSDKTestProjectID)
	sdkTestApplicationName = os.Getenv(EnvSynpseSDKTestApplicationName)
}

// getTestingClient returns a new API client for testing purposes. This
// client should be using project access keys.
func getTestingProjectClient(t *testing.T) *API {
	apiEndpointURL := os.Getenv(EnvSynpseAPIEndpointURL)
	accessKey := os.Getenv(EnvSynpseProjectAccessKey)
	projectID := os.Getenv(EnvSynpseSDKTestProjectID)

	var opts []Option
	if apiEndpointURL != "" {
		opts = append(opts, WithAPIEndpointURL(apiEndpointURL))
	}

	apiClient, err := NewWithProject(accessKey, projectID, opts...)
	require.NoError(t, err, "failed to create API client")

	return apiClient
}

// getTestingPersonalClient returns a new API client for testing purposes. This
// client should be using personal access keys.
func getTestingPersonalClient(t *testing.T) *API {
	apiEndpointURL := os.Getenv(EnvSynpseAPIEndpointURL)
	accessKey := os.Getenv(EnvSynpsePersonalAccessKey)
	projectID := os.Getenv(EnvSynpseSDKTestProjectID)

	var opts []Option
	if apiEndpointURL != "" {
		opts = append(opts, WithAPIEndpointURL(apiEndpointURL))
	}

	apiClient, err := NewWithProject(accessKey, projectID, opts...)
	require.NoError(t, err, "failed to create API client")

	return apiClient
}
