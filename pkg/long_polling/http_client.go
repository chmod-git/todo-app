package long_polling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AuthResponse struct {
	Token string `json:"token"`
}

type HTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *HTTPClient) Request(method, endpoint string, headers map[string]string, body interface{}) ([]byte, int, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)

	var requestBody []byte
	var err error
	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal body: %v", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %v", err)
	}

	return responseBody, resp.StatusCode, nil
}

func (c *HTTPClient) GET(endpoint string, headers map[string]string) ([]byte, int, error) {
	return c.Request(http.MethodGet, endpoint, headers, nil)
}

func (c *HTTPClient) POST(endpoint string, headers map[string]string, body interface{}) ([]byte, int, error) {
	return c.Request(http.MethodPost, endpoint, headers, body)
}

func (c *HTTPClient) PUT(endpoint string, headers map[string]string, body interface{}) ([]byte, int, error) {
	return c.Request(http.MethodPut, endpoint, headers, body)
}

func (c *HTTPClient) DELETE(endpoint string, headers map[string]string) ([]byte, int, error) {
	return c.Request(http.MethodDelete, endpoint, headers, nil)
}
