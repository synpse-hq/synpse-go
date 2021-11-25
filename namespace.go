package synpse

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type ListNamespacesRequest struct{}

// ListNamespaces list namespaces in the current project
func (api *API) ListNamespaces(ctx context.Context, req *ListNamespacesRequest) ([]*Namespace, error) {
	resp, _, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL), nil)
	if err != nil {
		return nil, err
	}

	var result []*Namespace
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return result, nil
}

// CreateNamespace creates a new namespace in the current project.
func (api *API) CreateNamespace(ctx context.Context, namespace Namespace) (*Namespace, error) {
	resp, _, err := api.makeRequestContext(ctx, http.MethodPost, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL), namespace)
	if err != nil {
		return nil, err
	}

	var result Namespace
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

// GetNamespace returns namespace by name
func (api *API) GetNamespace(ctx context.Context, namespace string) (*Namespace, error) {
	resp, _, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace), nil)
	if err != nil {
		return nil, err
	}

	var result Namespace
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

// UpdateNamespace updates namespace
func (api *API) UpdateNamespace(ctx context.Context, namespace Namespace) (*Namespace, error) {
	if namespace.Name == "" {
		return nil, errors.New("namespace name is required")
	}

	resp, _, err := api.makeRequestContext(ctx, http.MethodPut, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace.Name), namespace)
	if err != nil {
		return nil, err
	}

	var result Namespace
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

// DeleteNamespace delete namespace and all applications, secrets within it
func (api *API) DeleteNamespace(ctx context.Context, namespace string) error {
	_, _, err := api.makeRequestContext(ctx, http.MethodDelete, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace), nil)
	return err
}

// Namespace is used to separate applications and secrets that reside in the same project. Applications
// from multiple namespaces can be deployed on the same devices.
type Namespace struct {
	ID               string           `json:"id" yaml:"id"`
	CreatedAt        time.Time        `json:"createdAt" yaml:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt" yaml:"updatedAt"`
	ProjectID        string           `json:"projectId" yaml:"projectId"`
	Name             string           `json:"name" yaml:"name"`
	Config           *NamespaceConfig `json:"config,omitempty" yaml:"config,omitempty"`
	ApplicationCount int              `json:"applicationCount,omitempty" yaml:"applicationCount,omitempty"` // Read only
}

type NamespaceConfig struct {
	RegistryAuth DockerAuth `json:"registryAuthentication,omitempty" yaml:"registryAuthentication,omitempty"`
}
