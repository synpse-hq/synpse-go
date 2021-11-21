package synpse

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/function61/holepunch-server/pkg/wsconnadapter"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

func (api *API) ListDevices(ctx context.Context, filters []string) ([]*Device, error) {
	// construct filter query
	f := ""
	total := len(filters)
	if total > 0 {
		for i := 0; i < total-1; i++ {
			f = f + filters[i] + "&"
		}
		f = f + filters[total-1]
	}

	resp, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, devicesURL+"?q="+f), nil)
	if err != nil {
		return nil, err
	}

	var result []*Device
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return result, nil
}

func (api *API) GetDevice(ctx context.Context, device string) (*Device, error) {
	resp, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, devicesURL, device+"?full"), nil)
	if err != nil {
		return nil, err
	}
	var d Device
	err = json.Unmarshal(resp, &d)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &d, nil
}

func (api *API) DeleteDevice(ctx context.Context, project, device string) error {
	_, err := api.makeRequestContext(ctx, http.MethodDelete, getURL(api.BaseURL, projectsURL, api.ProjectID, devicesURL, device), nil)
	return err
}

// UpdateDevice can update device name and desired version
func (api *API) UpdateDevice(ctx context.Context, device Device) (*Device, error) {
	resp, err := api.makeRequestContext(ctx, http.MethodPatch, getURL(api.BaseURL, projectsURL, api.ProjectID, devicesURL, device.ID), device)
	if err != nil {
		return nil, err
	}

	var d Device
	err = json.Unmarshal(resp, &d)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &d, nil
}

func (api *API) DeviceSSH(ctx context.Context, deviceID string) (net.Conn, error) {
	req, err := http.NewRequestWithContext(ctx, "", "", nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(api.APIAccessKey, "")

	wsConn, _, err := websocket.DefaultDialer.Dial(getWebsocketURL(api.BaseURL, projectsURL, api.ProjectID, devicesURL, deviceID, sshURL), req.Header)
	if err != nil {
		return nil, err
	}

	return wsconnadapter.New(wsConn), nil
}

func (api *API) DeviceConnect(ctx context.Context, deviceID, port, hostname string) (net.Conn, error) {
	u := getWebsocketURL(api.BaseURL, projectsURL, api.ProjectID, devicesURL, deviceID, connectURL)
	req, err := http.NewRequestWithContext(ctx, "", u, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("port", port)
	q.Add("hostname", hostname)

	req.URL.RawQuery = q.Encode()
	req.SetBasicAuth(api.APIAccessKey, "")

	wsConn, _, err := websocket.DefaultDialer.Dial(req.URL.String(), req.Header)
	if err != nil {
		return nil, err
	}

	return wsconnadapter.New(wsConn), nil
}

func (api *API) DeviceReboot(ctx context.Context, deviceID string) error {
	_, err := api.makeRequestContext(ctx, http.MethodPost, getURL(projectsURL, api.ProjectID, devicesURL, deviceID, rebootURL), []byte{})
	return err
}

type Device struct {
	ID        string    `json:"id" yaml:"id"`
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" yaml:"updatedAt"`
	ProjectID string    `json:"projectId" yaml:"projectId"`

	Name                 string            `json:"name" yaml:"name"`
	RegistrationTokenID  string            `json:"registrationTokenId" yaml:"registrationTokenId"`
	AgentSettings        AgentSettings     `json:"agentSettings" yaml:"agentSettings"`
	Info                 DeviceInfo        `json:"info" yaml:"info" gorm:"json"`
	LastSeenAt           time.Time         `json:"lastSeenAt" yaml:"lastSeenAt"`
	Status               DeviceStatus      `json:"status" yaml:"status"`
	Labels               map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty" yaml:"environmentVariables,omitempty"`

	Applications []*Application `json:"applications,omitempty" yaml:"applications,omitempty"`
}

type DeviceStatus string

const (
	DeviceStatusOnline  = DeviceStatus("online")
	DeviceStatusOffline = DeviceStatus("offline")
)

type AgentSettings struct {
	AgentLogLevel            string `json:"agentLogLevel" yaml:"agentLogLevel"`
	DesiredAgentVersion      string `json:"desiredAgentVersion" yaml:"desiredAgentVersion"`
	DesiredAgentVersionForce bool   `json:"-" yaml:"-"` // TODO: Placeholder only so we can wire it in if needed from controller side. If this set to true agent will ignore downgrade checks
}

type DeviceInfo struct {
	DeviceID      string     `json:"deviceId" yaml:"deviceId"`
	AgentVersion  string     `json:"agentVersion" yaml:"agentVersion"`
	AgentLogLevel string     `json:"agentLogLevel" yaml:"agentLogLevel"`
	IPAddress     string     `json:"ipAddress" yaml:"ipAddress"`
	Architecture  string     `json:"architecture" yaml:"architecture"`
	Hostname      string     `json:"hostname" yaml:"hostname"`
	OSRelease     OSRelease  `json:"osRelease" yaml:"osRelease"`
	Docker        DockerInfo `json:"docker" yaml:"docker"`

	CPUInfo CPUInfo `json:"cpuInfo" yaml:"cpuInfo"`
}

type CPUInfo struct {
	BrandName      string `json:"brandName" yaml:"brandName"`           // Brand name reported by the CPU
	VendorString   string `json:"vendorString" yaml:"vendorString"`     // Raw vendor string.
	PhysicalCores  int    `json:"physicalCores" yaml:"physicalCores"`   // Number of physical processor cores in your CPU. Will be 0 if undetectable.
	ThreadsPerCore int    `json:"threadsPerCore" yaml:"threadsPerCore"` // Number of threads per physical core. Will be 1 if undetectable.
	LogicalCores   int    `json:"logicalCores" yaml:"logicalCores"`     // Number of physical cores times threads that can run on each core through the use of hyperthreading. Will be 0 if undetectable.
	Family         int    `json:"family" yaml:"family"`                 // CPU family number
	Model          int    `json:"model" yaml:"model"`                   // CPU model number
	Hz             int64  `json:"hz" yaml:"hz"`                         // Clock speed, if known, 0 otherwise. Will attempt to contain base clock speed.
}

type DockerInfo struct {
	Version           string `json:"version" yaml:"version"`
	PrivilegedEnabled bool   `json:"privilegedEnabled" yaml:"privilegedEnabled"`
	BridgeIP          string `json:"bridgeIP" yaml:"bridgeIP"`
	Runtimes          string `json:"runtimes" yaml:"runtimes"`
	OSType            string `json:"osType" yaml:"osType"`
	Health            string `json:"health" yaml:"health"`
	HealthDescription string `json:"healthDescription" yaml:"healthDescription"`
}

type OSRelease struct {
	PrettyName string `json:"prettyName" yaml:"prettyName"`
	Name       string `json:"name" yaml:"name"`
	VersionID  string `json:"versionId" yaml:"versionId"`
	Version    string `json:"version" yaml:"version"`
	ID         string `json:"id" yaml:"id"`
	IDLike     string `json:"idLike" yaml:"idLike"`
}
