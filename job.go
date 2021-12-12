package synpse

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/function61/holepunch-server/pkg/wsconnadapter"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type ListJobsRequest struct {
	Namespace string `json:"namespace"`
}

func (api *API) ListJobs(ctx context.Context, req ListJobsRequest) ([]*Job, error) {
	if req.Namespace == "" {
		return nil, fmt.Errorf("namespace not selected")
	}

	resp, _, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, req.Namespace, jobsURL), nil)
	if err != nil {
		return nil, err
	}

	var result []*Job
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return result, nil
}

func (api *API) CreateJob(ctx context.Context, namespace string, job Job) (*Job, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace not selected")
	}

	resp, _, err := api.makeRequestContext(ctx, http.MethodPost, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, jobsURL), job)
	if err != nil {
		return nil, err
	}

	var result Job
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

func (api *API) UpdateJob(ctx context.Context, namespace string, j Job) (*Job, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace not selected")
	}

	var jobIdentifier string
	switch {
	case j.ID != "":
		jobIdentifier = j.ID
	case j.Name != "":
		jobIdentifier = j.Name
	default:
		return nil, fmt.Errorf("job ID or name must be specified")
	}

	resp, _, err := api.makeRequestContext(ctx, http.MethodPatch, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, jobsURL, jobIdentifier), j)
	if err != nil {
		return nil, err
	}

	var result Job
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

func (api *API) DeleteJob(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		return fmt.Errorf("namespace not selected")
	}

	if name == "" {
		return fmt.Errorf("name or ID not selected")
	}

	_, _, err := api.makeRequestContext(ctx, http.MethodDelete, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, jobsURL, name), nil)
	if err != nil {
		return err
	}
	return nil
}

func (api *API) GetJob(ctx context.Context, namespace, name string) (*Job, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace not selected")
	}

	if name == "" {
		return nil, fmt.Errorf("name or ID not selected")
	}

	resp, _, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, jobsURL, name), nil)
	if err != nil {
		return nil, err
	}

	var result Job
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

func (api *API) DeviceJobLogs(ctx context.Context, namespace, jobID string, opts LogsOpts) (net.Conn, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace not selected")
	}

	u := getWebsocketURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, jobsURL, jobID, logsURL, opts.Container)
	req, err := http.NewRequestWithContext(ctx, "", u, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("device", opts.Device)
	q.Add("follow", strconv.FormatBool(opts.Follow))
	q.Add("tail", strconv.Itoa(opts.Tail))

	req.URL.RawQuery = q.Encode()
	req.SetBasicAuth(api.APIAccessKey, "")

	wsConn, _, err := websocket.DefaultDialer.Dial(req.URL.String(), req.Header)
	if err != nil {
		return nil, err
	}

	return wsconnadapter.New(wsConn), nil
}

// Job is a one-off run of the container run, similar to k8s job or `docker run` without daemonization. For jobs,
// restart policy is always - "no", meaning that the container will not be restarted if it fails or finishes.
type Job struct {
	// Primary ID
	ID string `json:"id" yaml:"id"`

	/* Fields set by the user */

	Name         string         `json:"name" yaml:"name"`
	Description  string         `json:"description,omitempty" yaml:"description,omitempty"`
	Scheduling   Scheduling     `json:"scheduling" yaml:"scheduling"`
	Spec         JobSpec        `json:"spec" yaml:"spec" validate:"dive"`
	DesiredState DeviceJobState `json:"desiredState" yaml:"desiredState"`

	/* Fields set by the server */

	Version       int64          `json:"version" yaml:"version"`             // Version is used to check application updated
	ConfigVersion int64          `json:"configVersion" yaml:"configVersion"` // config version is used on agent side only to indicate that the config has changed and application object needs to be redeployed.
	CreatedAt     time.Time      `json:"createdAt" yaml:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt" yaml:"updatedAt"`
	CompletedAt   time.Time      `json:"completedAt" yaml:"completedAt"` // last job instance on a device that finished running
	ProjectID     string         `json:"projectId" yaml:"projectId"`
	NamespaceID   string         `json:"namespaceId" yaml:"namespaceId"`
	State         DeviceJobState `json:"status" yaml:"status"` // Compute from device jobs on status updates
	DeviceJobs    []*DeviceJob   `json:"deviceJobs,omitempty" yaml:"deviceJobs,omitempty"`
}

type JobSpec struct {
	// Specs are parsed based on Type in main job struct
	ContainerSpec []ContainerSpec `json:"containers,omitempty" yaml:"containers,omitempty" validate:"dive"`
}

// DeviceJob represents a single instance of the job running on a device. This should have a start time, individual
// container statuses and the completion.
//
// When Job is created, scheduler creates a DeviceJob that is then being tracked by the individual device (scheduled via
// device bundle). Device then updates the job statuses.
type DeviceJob struct {
	ID           string           `json:"id" yaml:"id"`
	CreatedAt    time.Time        `json:"createdAt" yaml:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt" yaml:"updatedAt"`
	DeviceID     string           `json:"deviceId" yaml:"deviceId"`
	JobID        string           `json:"jobId" yaml:"jobId"`
	ProjectID    string           `json:"projectId" yaml:"projectId"`
	NamespaceID  string           `json:"namespaceId" yaml:"namespaceId""`
	DesiredState DeviceJobState   `json:"desiredState" yaml:"desiredState"`
	State        DeviceJobState   `json:"state" yaml:"state"`
	StartedAt    time.Time        `json:"startedAt" yaml:"startedAt"`
	CompletedAt  time.Time        `json:"completedAt" yaml:"completedAt"`
	Statuses     WorkloadStatuses `json:"statuses" yaml:"statuses"` // Reported by the device
}

// DeviceJobState is an overall state of the job. If at least a single underlying container is in a failed state, then
// the job is considered failed.
type DeviceJobState string

const (
	DeviceJobStatePending   DeviceJobState = "pending"
	DeviceJobStateRunning   DeviceJobState = "running"
	DeviceJobStateFailed    DeviceJobState = "failed"
	DeviceJobStateSucceeded DeviceJobState = "succeeded"
	DeviceJobStateStopped   DeviceJobState = "stopped"
)
