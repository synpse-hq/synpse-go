package synpse

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// CreateSecret creates a secret in a specified namespace.
// Secrets API ref: https://docs.synpse.net/synpse-core/applications/secrets
func (api *API) CreateSecret(ctx context.Context, namespace string, secret Secret) (*Secret, error) {
	if namespace == "" {
		return nil, ErrNamespaceNotSpecified
	}

	err := api.ensureSecretEncoding(&secret)
	if err != nil {
		return nil, fmt.Errorf("failed to encode secret payload: %w", err)
	}

	resp, err := api.makeRequestContext(ctx, http.MethodPost, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, secretsURL), secret)
	if err != nil {
		return nil, err
	}

	var result Secret
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

func (api *API) ensureSecretEncoding(secret *Secret) error {
	// Checking if string is encoded
	_, err := base64.StdEncoding.DecodeString(secret.Data)
	if err == nil {
		// Already encoded, nothing to do
		return nil
	}

	// Encode the payload
	secret.Data = base64.StdEncoding.EncodeToString([]byte(secret.Data))

	return nil
}

// DeleteSecret deletes secret from the namespace
func (api *API) DeleteSecret(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		return ErrNamespaceNotSpecified
	}

	if name == "" {
		return fmt.Errorf("secret name or ID not specified")
	}

	_, err := api.makeRequestContext(ctx, http.MethodDelete, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, secretsURL, name), nil)
	if err != nil {
		return err
	}

	return nil
}

// ListSecrets lists all secrets in a namespace
func (api *API) ListSecrets(ctx context.Context, namespace string) ([]*Secret, error) {
	if namespace == "" {
		return nil, ErrNamespaceNotSpecified
	}

	resp, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, secretsURL), nil)
	if err != nil {
		return nil, err
	}

	var result []*Secret
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return result, nil
}

// UpdateSecret updates a secret data in a specified namespace. Note that secret type cannot be changed.
func (api *API) UpdateSecret(ctx context.Context, namespace string, p Secret) (*Secret, error) {
	if namespace == "" {
		return nil, ErrNamespaceNotSpecified
	}

	if p.Name == "" {
		return nil, fmt.Errorf("secret name not specified")
	}

	err := api.ensureSecretEncoding(&p)
	if err != nil {
		return nil, fmt.Errorf("failed to encode secret payload: %w", err)
	}

	resp, err := api.makeRequestContext(ctx, http.MethodPatch, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, secretsURL, p.Name), p)
	if err != nil {
		return nil, err
	}

	var result Secret
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

func (api *API) GetSecret(ctx context.Context, namespace, name string) (*Secret, error) {
	if namespace == "" {
		return nil, ErrNamespaceNotSpecified
	}

	if name == "" {
		return nil, fmt.Errorf("secret name not specified")
	}

	resp, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, secretsURL, name), nil)
	if err != nil {
		return nil, err
	}

	var result Secret
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	decodedData, err := base64.StdEncoding.DecodeString(result.Data)
	if err == nil {
		result.Data = string(decodedData)
	}

	return &result, nil
}

// Secret is used to conceal sensitive configuration from the deployment manifests. You can create
// multiple secrets per namespace and use them across one or more applications. Environment type secrets
// can be used for applications as environment variables or also can be used in Docker registry authentication.
type Secret struct {
	ID          string     `json:"id,omitempty" yaml:"id,omitempty"`
	CreatedAt   time.Time  `json:"createdAt,omitempty" yaml:"createdAt,omitempty"`
	UpdatedAt   time.Time  `json:"updatedAt,omitempty" yaml:"updatedAt,omitempty"`
	Name        string     `json:"name" yaml:"name"`
	ProjectID   string     `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	NamespaceID string     `json:"namespaceId,omitempty" yaml:"namespaceId,omitempty"`
	Version     int64      `json:"version,omitempty" yaml:"version,omitempty"`
	Type        SecretType `json:"type,omitempty" yaml:"type,omitempty"`
	Data        string     `json:"data,omitempty" yaml:"data,omitempty"` // Base64 encoded data
}

type SecretType string

const (
	SecretTypeEnvironment SecretType = "Environment" // Can be used as environment variables
	SecretTypeFile        SecretType = "File"        // Can be mounted as a file to a container
)
