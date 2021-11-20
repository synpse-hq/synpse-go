package synpse

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

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

	DeviceCount int `json:"deviceCount" yaml:"deviceCount"`
}

type DeviceNamingStrategy struct {
	Type DeviceNamingStrategyType `json:"type" yaml:"type"`
}

type DeviceNamingStrategyType string

const (
	DeviceNamingStrategyTypeDefault      DeviceNamingStrategyType = "default"
	DeviceNamingStrategyTypeFromHostname DeviceNamingStrategyType = "fromHostname"
)
