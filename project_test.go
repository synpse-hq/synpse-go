package synpse

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListProjects(t *testing.T) {
	client := getTestingPersonalClient(t)

	projects, err := client.ListProjects(context.Background())
	require.NoError(t, err)

	assert.True(t, len(projects) > 0)

	var projectFound bool

	for _, project := range projects {
		if project.ID == sdkTestProjectID {
			assert.Equal(t, sdkTestProjectName, project.Name)
			projectFound = true
		}
	}

	assert.True(t, projectFound, "expected to find testing project, check SYNPSE_SDK_TEST_PROJECT_ID and SYNPSE_SDK_TEST_PROJECT_NAME env vars")
}
