package synpse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type ListApplicationAllocationsRequest struct {
	Namespace   string
	Application string // Name or ID
}

func (api *API) ListApplicationAllocations(ctx context.Context, req *ListApplicationAllocationsRequest) ([]*DeviceApplicationStatus, error) {
	if req.Namespace == "" {
		return nil, fmt.Errorf("namespace not selected")
	}

	resp, _, err := api.makeRequestContext(ctx, http.MethodGet, getURL(
		api.BaseURL,
		projectsURL, api.ProjectID, namespacesURL, req.Namespace, applicationsURL, req.Application, allocationsURL), nil)
	if err != nil {
		return nil, err
	}

	var applicationAllocs []*DeviceApplicationStatus

	err = json.Unmarshal(resp, &applicationAllocs)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return applicationAllocs, nil
}

// DeviceApplicationStatus represent application status running on a specific device
type DeviceApplicationStatus struct {
	DeviceID            string           `json:"deviceId" yaml:"deviceId"`
	DeviceName          string           `json:"deviceName" yaml:"deviceName"`
	ApplicationID       string           `json:"applicationId" yaml:"applicationId"`
	ProjectID           string           `json:"projectId" yaml:"projectId"`
	NamespaceID         string           `json:"namespaceId" yaml:"namespaceId"`
	LastSeen            time.Time        `json:"lastSeen" yaml:"lastSeen"`
	ApplicationStatuses []WorkloadStatus `json:"applicationStatuses" yaml:"applicationStatus"`
}
