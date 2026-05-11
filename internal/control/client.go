package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

var ErrRuntimeNotFound = errors.New("local Axis runtime not found")

type ClientError struct {
	Code    string
	Message string
	Hint    string
	Status  int
}

func (e *ClientError) Error() string {
	if e.Hint != "" {
		return e.Message + ". " + e.Hint
	}
	return e.Message
}

func (e *ClientError) Is(target error) bool {
	return target == ErrRuntimeNotFound && e.Code == "runtime_not_found"
}

type Client struct {
	locator    *RuntimeLocator
	httpClient *http.Client
}

func NewClient(locator *RuntimeLocator, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{locator: locator, httpClient: httpClient}
}

func (c *Client) SubmitTask(ctx context.Context, task *types.AgentTask) (SubmitTaskResponse, error) {
	var resp SubmitTaskResponse
	err := c.do(ctx, http.MethodPost, "/v1/tasks", SubmitTaskRequest{Task: task}, &resp)
	return resp, err
}

func (c *Client) Status(ctx context.Context, taskID string) (StatusResponse, error) {
	var resp StatusResponse
	err := c.do(ctx, http.MethodGet, "/v1/tasks/"+url.PathEscape(taskID)+"/status", nil, &resp)
	return resp, err
}

func (c *Client) Health(ctx context.Context) (HealthResponse, error) {
	var resp HealthResponse
	err := c.do(ctx, http.MethodGet, "/v1/health", nil, &resp)
	return resp, err
}

func (c *Client) do(ctx context.Context, method string, path string, body any, out any) error {
	record, err := c.locator.Load()
	if err != nil {
		if errors.Is(err, ErrRuntimeLocatorNotFound) {
			return &ClientError{Code: "runtime_not_found", Message: "No local Axis runtime found", Hint: "Start one with: axis start"}
		}
		return err
	}
	url := strings.TrimRight(record.Address, "/") + path
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return &ClientError{Code: "request_failed", Message: fmt.Sprintf("control request failed with status %d", resp.StatusCode), Status: resp.StatusCode}
		}
		return &ClientError{Code: errResp.Code, Message: errResp.Message, Hint: errResp.Hint, Status: resp.StatusCode}
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
