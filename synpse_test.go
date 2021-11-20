package synpse

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	EnvSynpseAccessKey = "SYNPSE_ACCESS_KEY"
	EnvSynpseProjectID = "SYNPSE_PROJECT_ID"
)

func getTestingClient(t *testing.T) *API {
	accessKey := os.Getenv(EnvSynpseAccessKey)
	projectID := os.Getenv(EnvSynpseProjectID)

	apiClient, err := New(accessKey, projectID)
	require.NoError(t, err, "failed to create API client")

	return apiClient
}
