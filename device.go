package synpse

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/function61/holepunch-server/pkg/wsconnadapter"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type ListDevicesRequest struct {
	SearchQuery       string            // Search query, e.g. "power-plant-one"
	Labels            map[string]string // A map of labels to match
	PaginationOptions PaginationOptions
}

type ListDevicesResponse struct {
	Devices    []*Device
	Pagination Pagination
}

func (api *API) ListDevices(ctx context.Context, req *ListDevicesRequest) (*ListDevicesResponse, error) {
	apiURL := getURL(api.BaseURL, projectsURL, api.ProjectID, devicesURL)

	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare URL '%s', error: %w", apiURL, err)
	}

	q := u.Query()

	if req.SearchQuery != "" {
		q.Add("q", req.SearchQuery)
	}

	if len(req.Labels) > 0 {
		bts, err := json.Marshal(req.Labels)
		if err != nil {
			return nil, fmt.Errorf("failed to encode labels, error: %w", err)
		}
		q.Add("labels", string(bts))
	}

	// Setting pagination
	setPagination(q, &req.PaginationOptions)

	u.RawQuery = q.Encode()

	// resp, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, devicesURL+"?q="+f), nil)
	resp, respHeader, err := api.makeRequestContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	var result []*Device
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return &ListDevicesResponse{
		Devices:    result,
		Pagination: getPagination(respHeader),
	}, nil
}

func (api *API) GetDevice(ctx context.Context, device string) (*Device, error) {
	resp, _, err := api.makeRequestContext(ctx, http.MethodGet, getURL(api.BaseURL, projectsURL, api.ProjectID, devicesURL, device+"?full"), nil)
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

func (api *API) DeleteDevice(ctx context.Context, device string) error {
	_, _, err := api.makeRequestContext(ctx, http.MethodDelete, getURL(api.BaseURL, projectsURL, api.ProjectID, devicesURL, device), nil)
	return err
}

// UpdateDevice can update device name and desired version
func (api *API) UpdateDevice(ctx context.Context, device Device) (*Device, error) {
	resp, _, err := api.makeRequestContext(ctx, http.MethodPatch, getURL(api.BaseURL, projectsURL, api.ProjectID, devicesURL, device.ID), device)
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

func (api *API) GetDeviceSSHClient(ctx context.Context, deviceID string) (*ssh.Client, error) {
	req, err := http.NewRequestWithContext(ctx, "", "", nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(api.APIAccessKey, "")

	wsConn, _, err := websocket.DefaultDialer.Dial(getWebsocketURL(api.BaseURL, projectsURL, api.ProjectID, devicesURL, deviceID, sshURL), req.Header)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the backend: %w", err)
	}

	deviceConn := wsconnadapter.New(wsConn)

	sshClient, err := api.newSSHClient(ctx, deviceConn)
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection, error: %w", err)
	}

	return sshClient, nil
}

func (api *API) RunDeviceCommand(ctx context.Context, deviceID, command string) (string, error) {
	sshClient, err := api.GetDeviceSSHClient(ctx, deviceID)
	if err != nil {
		return "", err
	}

	session, err := sshClient.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	out, err := session.CombinedOutput(command)
	if err != nil {
		return string(out), fmt.Errorf("SSH command error: %w", err)
	}
	return string(out), nil
}

func (api *API) newSSHClient(ctx context.Context, deviceConn net.Conn) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		Timeout:         time.Second * 5,
		User:            "",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	config.Auth = []ssh.AuthMethod{}

	c, err := newSSHClientConn(deviceConn, ":0", config)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client conn: %w", err)
	}

	go func() {
		err := sshKeepAlive(ctx, c, deviceConn)
		if err != nil {
			api.logger.Printf("Failed to setup SSH client keepalives: %s", err)
		}
	}()

	return c, nil
}

func newSSHClientConn(conn net.Conn, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

func sshKeepAlive(ctx context.Context, cl *ssh.Client, conn net.Conn) error {
	const keepAliveInterval = 18 * time.Second
	t := time.NewTicker(keepAliveInterval)
	defer t.Stop()
	for {
		deadline := time.Now().Add(keepAliveInterval).Add(15 * time.Second)
		err := conn.SetDeadline(deadline)
		if err != nil {
			return errors.Wrap(err, "failed to set deadline")
		}
		select {
		case <-t.C:
			_, _, err = cl.SendRequest("keepalive@synpse.net", true, nil)
			if err != nil {
				return errors.Wrap(err, "failed to send keep alive")
			}
		case <-ctx.Done():
			return nil
		}
	}
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
	_, _, err := api.makeRequestContext(ctx, http.MethodPost, getURL(projectsURL, api.ProjectID, devicesURL, deviceID, rebootURL), []byte{})
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
