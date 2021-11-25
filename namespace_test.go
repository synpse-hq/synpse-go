package synpse

import (
	"context"
	"testing"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testNamespacePrefix = "sdk-ns-test-"

func TestNamespaces(t *testing.T) {
	ctx := context.Background()
	client := getTestingProjectClient(t)

	nsName := testNamespacePrefix + ksuid.New().String()

	_, err := client.ListNamespaces(ctx, &ListNamespacesRequest{})
	require.NoError(t, err)

	namespace, err := client.CreateNamespace(ctx, Namespace{
		Name: nsName,
	})
	require.NoError(t, err, "failed to create namespace")

	t.Cleanup(func() {
		err := client.DeleteNamespace(ctx, namespace.Name)
		assert.NoError(t, err, "failed to delete namespace")
	})

	assert.Equal(t, nsName, namespace.Name, "namespace name doesn't match")

	namespaces, err := client.ListNamespaces(ctx, &ListNamespacesRequest{})
	require.NoError(t, err)

	assert.Contains(t, namespaces, namespace, "namespace not found")
}
