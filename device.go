package synpse

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

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
	resp, err := api.makeRequestContext(ctx, http.MethodGet, getURL(projectsURL, api.ProjectID, devicesURL, device+"?full"), nil)
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
	_, err := api.makeRequestContext(ctx, http.MethodDelete, getURL(projectsURL, api.ProjectID, devicesURL, device), nil)
	return err
}

// UpdateDevice can update device name and desired version
func (api *API) UpdateDevice(ctx context.Context, device Device) (*Device, error) {
	resp, err := api.makeRequestContext(ctx, http.MethodPatch, getURL(projectsURL, api.ProjectID, devicesURL, device.ID), device)
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
