package client

import (
	"bytes"
	"net/http"
	"net/url"
)

func (c *client) doRequest(urlSuffix string, httpMethod string, requestBody []byte) (*http.Response, error) {
	url, err := url.JoinPath(c.serverAddr, urlSuffix)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(requestBody)
	req, err := http.NewRequest(httpMethod, url, buffer)
	if err != nil {
		return nil, err
	}

	if len(c.token) != 0 {
		req.Header.Set("X-Auth-Token", c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
