package synpse

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListApplications(t *testing.T) {
	client := getTestingProjectClient(t)

	applications, err := client.ListApplications(context.Background(), sdkTestNamespace)
	require.NoError(t, err)

	applicationFound := false

	for _, application := range applications {
		if application.Name == sdkTestApplicationName {
			applicationFound = true
			break
		}
	}
	require.True(t, applicationFound, "application not found")
}
