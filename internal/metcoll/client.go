package metcoll

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ArtemShalinFe/metcoll/internal/metrics"
)

type Logger interface {
	Infof(template string, args ...interface{})
}

type Client struct {
	host       string
	httpClient *http.Client
	logger     Logger
}

func NewClient(Host string, logger Logger) *Client {

	return &Client{
		host:       Host,
		httpClient: &http.Client{},
		logger:     logger,
	}

}

func (c *Client) prepareRequest(body []byte, url string) (*http.Request, error) {

	var zBuf bytes.Buffer
	zw := gzip.NewWriter(&zBuf)

	if _, err := zw.Write(body); err != nil {
		return nil, fmt.Errorf("cannot write compressed body err: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("cannot close compress writer err: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, &zBuf)
	if err != nil {
		return nil, fmt.Errorf("cannot create request err: %w", err)
	}
	req.Close = true

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	return req, nil
}

func (c *Client) Update(metric *metrics.Metrics) error {

	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("cannot marshal metric err: %w", err)
	}

	url, err := url.JoinPath("http://", c.host, "/update/")
	if err != nil {
		return fmt.Errorf("cannot join elements in path err: %w", err)
	}

	req, err := c.prepareRequest(body, url)
	if err != nil {
		return fmt.Errorf("cannot prepare request err: %w", err)
	}

	return c.DoRequest(req)

}

func (c *Client) BatchUpdate(metrics []*metrics.Metrics) error {

	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("cannot marshal metric err: %w", err)
	}

	url, err := url.JoinPath("http://", c.host, "/updates/")
	if err != nil {
		return fmt.Errorf("cannot join elements in path err: %w", err)
	}

	req, err := c.prepareRequest(body, url)
	if err != nil {
		return fmt.Errorf("cannot prepare request err: %w", err)
	}

	return c.DoRequest(req)

}

func (c *Client) DoRequest(req *http.Request) error {

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request execute err: %w", err)
	}

	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body err: %w", err)
	}

	c.logger.Infof(`request for update metric has been completed
	code: %d
	result: %s`, resp.StatusCode, string(res))

	return nil

}
