package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPError struct {
	Method     string
	Path       string
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	msg := strings.TrimSpace(string(e.Body))
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	return fmt.Sprintf("%s %s failed: HTTP %d: %s", e.Method, e.Path, e.StatusCode, msg)
}

type Client struct {
	baseURL       *url.URL
	authHeaderKey string
	authHeaderVal string
	httpClient    *http.Client
}

func NewClient(baseURL, token string, timeout time.Duration) (*Client, error) {
	return NewClientWithAuth(baseURL, "Authorization", bearerToken(token), timeout)
}

func NewClientWithAuth(baseURL, authHeaderKey, authHeaderVal string, timeout time.Duration) (*Client, error) {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid base URL: %q", baseURL)
	}
	u.Path = strings.TrimRight(u.Path, "/")

	return &Client{
		baseURL:       u,
		authHeaderKey: authHeaderKey,
		authHeaderVal: authHeaderVal,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func bearerToken(token string) string {
	if token == "" {
		return ""
	}
	return "Bearer " + token
}

func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body []byte) ([]byte, int, http.Header, error) {
	u := *c.baseURL
	u.Path = strings.TrimRight(c.baseURL.Path, "/") + path
	u.RawQuery = query.Encode()

	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), r)
	if err != nil {
		return nil, 0, nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.authHeaderKey != "" && c.authHeaderVal != "" {
		req.Header.Set(c.authHeaderKey, c.authHeaderVal)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, resp.Header, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return data, resp.StatusCode, resp.Header, &HTTPError{
			Method:     method,
			Path:       path,
			StatusCode: resp.StatusCode,
			Body:       data,
		}
	}
	return data, resp.StatusCode, resp.Header, nil
}

func JSON(data []byte) []byte {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return data
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return data
	}
	out, err := json.Marshal(v)
	if err != nil {
		return data
	}
	return append(out, '\n')
}

func PrettyJSON(data []byte) []byte {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return data
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return data
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return data
	}
	return append(out, '\n')
}
