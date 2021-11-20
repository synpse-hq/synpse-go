package synpse

import (
	"context"
	"testing"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

const testAppPrefix = "sdk-test-"

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

func TestCreateApplication(t *testing.T) {
	client := getTestingProjectClient(t)

	applicationName := testAppPrefix + ksuid.New().String()

	application, err := client.CreateApplication(context.Background(), sdkTestNamespace, Application{
		Name:        applicationName,
		Description: "test-app",
		Scheduling: Scheduling{
			Type: ScheduleTypeNoDevices,
		},
		Spec: ApplicationSpec{
			ContainerSpec: []ContainerSpec{
				{
					Name:  "hello",
					Image: "quay.io/synpse/hello-synpse-go",
					Ports: []string{"8080:8080"},
				},
			},
		},
	})
	require.NoError(t, err, "failed to create application")

	t.Logf("created application %s (%s)", application.Name, application.ID)

	t.Run("FindCreatedApplication", func(t *testing.T) {
		applications, err := client.ListApplications(context.Background(), sdkTestNamespace)
		require.NoError(t, err)

		applicationFound := false
		for _, application := range applications {
			if application.Name == applicationName {
				applicationFound = true
				break
			}
		}
		require.True(t, applicationFound, "created application not found")
	})

	t.Run("CheckApplicationDetails", func(t *testing.T) {
		stored, err := client.GetApplication(context.Background(), sdkTestNamespace, applicationName)
		require.NoError(t, err)

		require.Equal(t, application.ID, stored.ID)
		require.Equal(t, application.Name, stored.Name)
		require.Equal(t, application.Description, stored.Description)
		require.Equal(t, application.Scheduling.Type, stored.Scheduling.Type)
		require.Equal(t, application.Spec.ContainerSpec[0].Name, stored.Spec.ContainerSpec[0].Name)
		require.Equal(t, application.Spec.ContainerSpec[0].Image, stored.Spec.ContainerSpec[0].Image)
		require.Equal(t, application.Spec.ContainerSpec[0].Ports[0], stored.Spec.ContainerSpec[0].Ports[0])
	})

	// Delete application
	t.Run("DeleteApplication", func(t *testing.T) {
		err := client.DeleteApplication(context.Background(), sdkTestNamespace, applicationName)
		require.NoError(t, err)

		_, err = client.GetApplication(context.Background(), sdkTestNamespace, applicationName)
		require.Error(t, err, "expected to get an error")
	})
}
