package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gpufleet/internal/auth"
	"gpufleet/internal/model"
)

type Client struct {
	ServerURL string
	DeviceID  string
	Secret    string
	Timeout   time.Duration
	UseGzip   bool
	HTTP      *http.Client
}

type HTTPStatusError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *HTTPStatusError) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Body) == "" {
		return fmt.Sprintf("server returned %s", e.Status)
	}
	return fmt.Sprintf("server returned %s: %s", e.Status, strings.TrimSpace(e.Body))
}

func NewHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		clone := transport.Clone()
		clone.IdleConnTimeout = 30 * time.Second
		clone.ResponseHeaderTimeout = timeout
		client.Transport = clone
	}
	return client
}

func (c *Client) PostHeartbeat(heartbeat model.Heartbeat) error {
	return c.postJSON("/api/v1/agent/heartbeat", heartbeat)
}

func (c *Client) PostSamples(batch model.SampleBatch) error {
	return c.postJSON("/api/v1/agent/samples", batch)
}

func (c *Client) PostProcesses(batch model.ProcessBatch) error {
	return c.postJSON("/api/v1/agent/process-snapshots", batch)
}

func (c *Client) PostConfig(report model.AgentConfigReport) error {
	return c.postJSON("/api/v1/agent/config", report)
}

func (c *Client) GetUpdatePolicy() (model.AgentUpdatePolicy, error) {
	var response struct {
		Policy model.AgentUpdatePolicy `json:"policy"`
	}
	if err := c.postJSONDecode("/api/v1/agent/update-policy", map[string]string{
		"agent_version": model.AgentVersion,
	}, &response); err != nil {
		return model.AgentUpdatePolicy{}, err
	}
	return response.Policy, nil
}

func (c *Client) PostUpdateEvent(event model.AgentUpdateEvent) error {
	return c.postJSON("/api/v1/agent/update-events", event)
}

func (c *Client) postJSON(path string, value any) error {
	return c.postJSONDecode(path, value, nil)
}

func (c *Client) postJSONDecode(path string, value any, out any) error {
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	requestBody := body
	var reader io.Reader = bytes.NewReader(requestBody)
	contentEncoding := ""
	if c.UseGzip {
		var compressed bytes.Buffer
		gw := gzip.NewWriter(&compressed)
		if _, err := gw.Write(body); err != nil {
			_ = gw.Close()
			return err
		}
		if err := gw.Close(); err != nil {
			return err
		}
		requestBody = compressed.Bytes()
		reader = bytes.NewReader(requestBody)
		contentEncoding = "gzip"
	}

	endpoint, err := joinURL(c.ServerURL, path)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if contentEncoding != "" {
		req.Header.Set("Content-Encoding", contentEncoding)
	}
	if err := auth.AttachSignedHeaders(req, body, c.DeviceID, c.Secret, time.Now().UTC()); err != nil {
		return err
	}

	client := c.HTTP
	if client == nil {
		client = NewHTTPClient(timeout)
	}
	res, err := client.Do(req)
	if err != nil {
		client.CloseIdleConnections()
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		statusErr := &HTTPStatusError{
			StatusCode: res.StatusCode,
			Status:     res.Status,
			Body:       strings.TrimSpace(string(limited)),
		}
		if retryableHTTPStatus(res.StatusCode) {
			client.CloseIdleConnections()
		}
		return statusErr
	}
	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return err
		}
	}
	return nil
}

func joinURL(base, path string) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + path
	return parsed.String(), nil
}

func retryableHTTPStatus(status int) bool {
	return status == http.StatusRequestTimeout ||
		status == http.StatusTooManyRequests ||
		status == http.StatusInsufficientStorage ||
		status >= 500
}

func permanentQueueUploadError(err error) bool {
	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) {
		return false
	}
	return statusErr.StatusCode == http.StatusBadRequest ||
		statusErr.StatusCode == http.StatusRequestEntityTooLarge
}
