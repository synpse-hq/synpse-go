package synpse

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

func (c *API) ListDevices(ctx context.Context, project string, filters []string) ([]*Device, error) {

	// construct filter query
	f := ""
	total := len(filters)
	if total > 0 {
		for i := 0; i < total-1; i++ {
			f = f + filters[i] + "&"
		}
		f = f + filters[total-1]
	}

	resp, err := c.makeRequestContext(ctx, http.MethodGet, getURL(c.BaseURL, projectsURL, project, devicesURL+"?q="+f), nil)
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
