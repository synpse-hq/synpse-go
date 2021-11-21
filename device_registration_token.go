package synpse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

func (api *API) ListDeviceRegistrationTokens(ctx context.Context) ([]*DeviceRegistrationToken, error) {
	resp, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, deviceRegistrationTokenURL), nil)
	if err != nil {
		return nil, err
	}

	var result []*DeviceRegistrationToken
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return result, nil
}

func (api *API) CreateRegistrationToken(ctx context.Context, registrationToken DeviceRegistrationToken) (*DeviceRegistrationToken, error) {
	resp, err := api.makeRequestContext(ctx, http.MethodPost, getURL(api.BaseURL, projectsURL, api.ProjectID, deviceRegistrationTokenURL), registrationToken)
	if err != nil {
		return nil, err
	}

	var result DeviceRegistrationToken
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

func (api *API) UpdateRegistrationToken(ctx context.Context, registrationToken DeviceRegistrationToken) (*DeviceRegistrationToken, error) {
	if registrationToken.ID == "" {
		return nil, fmt.Errorf("registration token ID not specified")
	}
	resp, err := api.makeRequestContext(ctx, http.MethodPut, getURL(api.BaseURL, projectsURL, api.ProjectID, deviceRegistrationTokenURL, registrationToken.ID), registrationToken)
	if err != nil {
		return nil, err
	}

	var result DeviceRegistrationToken
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

func (api *API) DeleteRegistrationToken(ctx context.Context, registrationToken string) error {
	if registrationToken == "" {
		return fmt.Errorf("registration token ID not specified")
	}
	_, err := api.makeRequestContext(ctx, http.MethodDelete, getURL(api.BaseURL, projectsURL, api.ProjectID, deviceRegistrationTokenURL, registrationToken), nil)
	return err
}

func (api *API) GetDefaultDeviceRegistrationToken(ctx context.Context, project string) (*DeviceRegistrationToken, error) {
	resp, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, deviceRegistrationTokenURL), nil)
	if err != nil {
		return nil, err
	}

	var result DeviceRegistrationToken
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}
	return &result, nil
}

type DeviceRegistrationToken struct {
	ID                   string               `json:"id" yaml:"id"`
	CreatedAt            time.Time            `json:"createdAt" yaml:"createdAt"`
	UpdatedAt            time.Time            `json:"updatedAt" yaml:"updatedAt"`
	ProjectID            string               `json:"projectId" yaml:"projectId"`
	MaxRegistrations     *int                 `json:"maxRegistrations" yaml:"maxRegistrations"`
	Name                 string               `json:"name" yaml:"name"`
	Description          string               `json:"description" yaml:"description"`
	Labels               map[string]string    `json:"labels" yaml:"labels"`
	EnvironmentVariables map[string]string    `json:"environmentVariables" yaml:"environmentVariables"`
	NamingStrategy       DeviceNamingStrategy `json:"namingStrategy" yaml:"namingStrategy"`

	DeviceCount int `json:"deviceCount" yaml:"deviceCount"` // read-only
}

type DeviceNamingStrategy struct {
	Type DeviceNamingStrategyType `json:"type" yaml:"type"`
}

type DeviceNamingStrategyType string

const (
	DeviceNamingStrategyTypeDefault      DeviceNamingStrategyType = "default"
	DeviceNamingStrategyTypeFromHostname DeviceNamingStrategyType = "fromHostname"
)
