package client

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

func (c *client) doRequest(urlSuffix string, httpMethod string, token string, requestBody []byte) ([]byte, error) {
	url, err := url.JoinPath(c.serverAddr, urlSuffix)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(requestBody)

	req, err := http.NewRequest(httpMethod, url, buffer)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-Token", c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
