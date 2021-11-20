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

// CreateApplication creates a new application in the specified namespace.
// Applications API ref: https://docs.synpse.net/synpse-core/applications
func (api *API) CreateApplication(ctx context.Context, namespace string, application Application) (*Application, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace not selected")
	}

	resp, err := api.makeRequestContext(ctx, http.MethodPost, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, applicationsURL), application)
	if err != nil {
		return nil, err
	}

	var result Application
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

func (api *API) DeleteApplication(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		return fmt.Errorf("namespace not selected")
	}

	_, err := api.makeRequestContext(ctx, http.MethodDelete, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, applicationsURL, name), nil)
	if err != nil {
		return err
	}

	return err
}

func (api *API) ListApplications(ctx context.Context, namespace string) ([]*Application, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace not selected")
	}

	resp, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, applicationsURL), nil)
	if err != nil {
		return nil, err
	}

	var applications []*Application
	err = json.Unmarshal(resp, &applications)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return applications, nil
}

// UpdateApplication updates application. This will trigger a version bump on the server side which will redeploy application on
// all devices that the application is scheduled on.
func (api *API) UpdateApplication(ctx context.Context, namespace string, p Application) (*Application, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace not selected")
	}

	resp, err := api.makeRequestContext(ctx, http.MethodPatch, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, applicationsURL, p.Name), p)
	if err != nil {
		return nil, err
	}

	var result Application
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

// GetApplication gets application by name.
func (api *API) GetApplication(ctx context.Context, namespace, name string) (*Application, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace not selected")
	}

	resp, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, applicationsURL, name), nil)
	if err != nil {
		return nil, err
	}

	var result Application
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &result, nil
}

// LogsOpts is the structure that needs to be passed when getting application logs from the device.
type LogsOpts struct {
	Container string
	Device    string
	Follow    bool
	Tail      int
}

func (api *API) DeviceApplicationLogs(ctx context.Context, namespace, applicationID string, opts LogsOpts) (net.Conn, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace not selected")
	}

	u := getWebsocketURL(api.BaseURL, projectsURL, api.ProjectID, namespacesURL, namespace, applicationsURL, applicationID, logsURL, opts.Container)
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

// Application is
type Application struct {
	ID            string      `json:"id" yaml:"id"`
	Version       int64       `json:"version" yaml:"version"`             // Version is used to check application updated
	ConfigVersion int64       `json:"configVersion" yaml:"configVersion"` // config version is used on agent side only to indicate that the config has changed and application object needs to be redeployed.
	CreatedAt     time.Time   `json:"createdAt" yaml:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt" yaml:"updatedAt"`
	ProjectID     string      `json:"projectId" yaml:"projectId"`
	NamespaceID   string      `json:"namespaceId" yaml:"namespaceId"`
	Name          string      `json:"name" yaml:"name"`
	Description   string      `json:"description,omitempty" yaml:"description,omitempty"`
	Type          RuntimeType `json:"type" yaml:"type"`
	Scheduling    Scheduling  `json:"scheduling" yaml:"scheduling"`

	Spec             ApplicationSpec             `json:"spec" yaml:"spec"`
	DeploymentStatus ApplicationDeploymentStatus `json:"deploymentStatus,omitempty" yaml:"deploymentStatus,omitempty"` // computed
}

type Scheduling struct {
	Type      ScheduleType      `json:"type"`
	Selectors map[string]string `json:"selectors"`
}

type ScheduleType string

const (
	ScheduleTypeNoDevices   = "NoDevices" // Optional, defaults when no type and no selectors are specified
	ScheduleTypeAllDevices  = "AllDevices"
	ScheduleTypeConditional = "Conditional" // Optional, defaults when no type but selectors are specified
)

type RuntimeType string

const (
	RuntimeContainer RuntimeType = "container"
	RuntimeSystemd   RuntimeType = "systemd"
)

type ApplicationSpec struct {
	// Specs are parsed based on Type in main application struct
	ContainerSpec []ContainerSpec `json:"containers,omitempty" yaml:"containers,omitempty" validate:"dive"`
}

type ContainerSpec struct {
	// Container runtime
	Name             string      `json:"name,omitempty" yaml:"name,omitempty" codec:"name" validate:"name"`
	Image            string      `json:"image,omitempty" yaml:"image,omitempty" codec:"image"`
	Args             []string    `json:"args,omitempty" yaml:"args,omitempty" codec:"args"`
	Auth             *DockerAuth `json:"auth,omitempty" yaml:"auth,omitempty" codec:"auth"`
	CapAdd           []string    `json:"capAdd,omitempty" yaml:"capAdd,omitempty" codec:"cap_add"`
	CapDrop          []string    `json:"capDrop,omitempty" yaml:"capDrop,omitempty" codec:"cap_drop"`
	Command          string      `json:"command,omitempty" yaml:"command,omitempty" codec:"command"`
	GPUs             string      `json:"gpus,omitempty" yaml:"gpus,omitempty" codec:"gpus"` // Shortcut flag for the GPUs, to enable all gpus, specify 'all'
	Entrypoint       []string    `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty" codec:"entrypoint"`
	ForcePull        bool        `json:"forcePull,omitempty" yaml:"forcePull,omitempty" codec:"force_pull"`
	Hostname         string      `json:"hostname,omitempty" yaml:"hostname,omitempty" codec:"hostname"`
	MemoryHardLimit  int64       `json:"memoryHardLimit,omitempty" yaml:"memoryHardLimit,omitempty" codec:"memory_hard_limit"`
	User             string      `json:"user,omitempty" yaml:"user,omitempty" codec:"user"`
	NetworkMode      NetworkMode `json:"networkMode,omitempty" yaml:"networkMode,omitempty" codec:"network_mode"`
	Ports            []string    `json:"ports,omitempty" yaml:"ports,omitempty" codec:"ports"` // Ports to expose like 8080:800
	Privileged       bool        `json:"privileged,omitempty" yaml:"privileged,omitempty" codec:"privileged"`
	ImagePullTimeout string      `json:"imagePullTimeout,omitempty" yaml:"imagePullTimeout,omitempty" codec:"image_pull_timeout"`
	SecurityOpt      []string    `json:"securityOpt,omitempty" yaml:"securityOpt,omitempty" codec:"security_opt"`
	ShmSize          int64       `json:"shmSize,omitempty" yaml:"shmSize,omitempty" codec:"shm_size"`
	Volumes          []string    `json:"volumes,omitempty" yaml:"volumes,omitempty" codec:"volumes"`
	VolumeDriver     string      `json:"volumeDriver,omitempty" yaml:"volumeDriver,omitempty" codec:"volume_driver"`
	WorkDir          string      `json:"workDir,omitempty" yaml:"workDir,omitempty" codec:"work_dir"`

	// synpse specific
	Environment   Environments  `json:"env,omitempty" yaml:"env,omitempty" validate:"dive"`
	Secrets       []SecretRef   `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	RestartPolicy RestartPolicy `json:"restartPolicy,omitempty" yaml:"restartPolicy,omitempty"`
}

type NetworkMode string

const (
	NetworkModeHost     = "host"
	NetworkModeIsolated = "isolated"
	NetworkModeBridge   = "bridge"
)

type Environment struct {
	Name       string `json:"name" yaml:"name"`
	Value      string `json:"value,omitempty" yaml:"value,omitempty"`
	FromSecret string `json:"fromSecret,omitempty" yaml:"fromSecret,omitempty"`
}

// ByAge implements sort.Interface for []Person based on
// the Age field.
type Environments []Environment

func (e Environments) Len() int           { return len(e) }
func (e Environments) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e Environments) Less(i, j int) bool { return e[i].Name < e[j].Name }

type SecretRef struct {
	Name string `json:"name" yaml:"name"`
	// Filepath specifies where the secret should be mounted
	// into a running container. This is used in the application spec itself and
	// not stored in the database
	Filepath string `json:"filepath,omitempty" yaml:"filepath,omitempty"`
}

type RestartPolicy struct {
	Name              string `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
	MaximumRetryCount int    `json:"maximumRetryCount,omitempty" yaml:"maximumRetryCount,omitempty" toml:"maximumRetryCount,omitempty"`
}

// AlwaysRestart returns a restart policy that tells the Docker daemon to
// always restart the container.
func AlwaysRestart() RestartPolicy {
	return RestartPolicy{Name: "always"}
}

// RestartOnFailure returns a restart policy that tells the Docker daemon to
// restart the container on failures, trying at most maxRetry times.
func RestartOnFailure(maxRetry int) RestartPolicy {
	return RestartPolicy{Name: "on-failure", MaximumRetryCount: maxRetry}
}

// RestartUnlessStopped returns a restart policy that tells the Docker daemon to
// always restart the container except when user has manually stopped the container.
func RestartUnlessStopped() RestartPolicy {
	return RestartPolicy{Name: "unless-stopped"}
}

// NeverRestart returns a restart policy that tells the Docker daemon to never
// restart the container on failures.
func NeverRestart() RestartPolicy {
	return RestartPolicy{Name: "no"}
}

type DockerAuth struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty" codec:"username"`
	// Leave empty to get password from secret
	Password   string `json:"password,omitempty" yaml:"password,omitempty" codec:"password"`
	Email      string `json:"email,omitempty" yaml:"email,omitempty" codec:"email"`
	ServerAddr string `json:"serverAddress,omitempty" yaml:"serverAddress,omitempty" codec:"serverAddress"`
	// FromSecret populates password from secret
	FromSecret string `json:"fromSecret,omitempty" yaml:"fromSecret,omitempty" codec:"fromSecret"`
}

// ApplicationDeploymentStatus - high level status of the application deployment
// progress, computed on-the-fly based on the stats from the application/device status
// and what the scheduler thinks should be deployed
type ApplicationDeploymentStatus struct {
	Pending   int `json:"pending" yaml:"pending"`
	Available int `json:"available" yaml:"available"`
	Total     int `json:"total" yaml:"total"`
}
