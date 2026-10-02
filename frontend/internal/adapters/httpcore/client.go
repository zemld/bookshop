package httpcore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	api    string
	client *http.Client
}

func New(api string) *Client {
	return &Client{
		api: strings.TrimRight(api, "/"),
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: http.DefaultTransport.(*http.Transport).Clone(),
		},
	}
}

func (c *Client) CloseIdleConnections() {
	c.client.CloseIdleConnections()
}

// Stop releases the transport owned by this client after inbound requests drain.
func (c *Client) Stop() { c.CloseIdleConnections() }

func (c *Client) Request(ctx context.Context, method, path string, input, output any) error {
	var body io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.api+path, body)
	if err != nil {
		return err
	}
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode >= 400 {
		var problem struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(res.Body).Decode(&problem)
		return fmt.Errorf("API %d: %s", res.StatusCode, problem.Error)
	}
	if output != nil {
		return json.NewDecoder(res.Body).Decode(output)
	}
	return nil
}
