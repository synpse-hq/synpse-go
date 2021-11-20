package synpse

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListDevices(t *testing.T) {
	client := getTestingClient(t)

	devices, err := client.ListDevices(context.Background(), []string{})
	require.NoError(t, err)

	assert.True(t, len(devices) > 0)
}
