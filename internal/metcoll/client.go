package metcoll

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
	"syscall"
	"time"

	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/hashicorp/go-retryablehttp"
)

type Client struct {
	host       string
	httpClient *retryablehttp.Client
	logger     *logger.AppLogger
}

func NewClient(Host string, logger *logger.AppLogger) *Client {

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 2
	retryClient.CheckRetry = CheckRetry
	retryClient.RetryWaitMin = time.Duration(3 * time.Second)
	retryClient.RetryWaitMax = time.Duration(5 * time.Second)
	retryClient.Logger = logger

	return &Client{
		host:       Host,
		httpClient: retryClient,
		logger:     logger,
	}

}

func CheckRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {

	if errors.Is(err, syscall.ECONNREFUSED) {
		return true, err
	} else {
		return false, err
	}

}

func (c *Client) prepareRequest(ctx context.Context, body []byte, url string) (*retryablehttp.Request, error) {

	var zBuf bytes.Buffer
	zw := gzip.NewWriter(&zBuf)

	if _, err := zw.Write(body); err != nil {
		return nil, fmt.Errorf("cannot write compressed body err: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("cannot close compress writer err: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, url, &zBuf)
	if err != nil {
		return nil, fmt.Errorf("cannot create request err: %w", err)
	}
	req.Close = true

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	return req, nil
}

func (c *Client) Update(ctx context.Context, metric *metrics.Metrics) error {

	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("cannot marshal metric err: %w", err)
	}

	url, err := url.JoinPath("http://", c.host, "/update/")
	if err != nil {
		return fmt.Errorf("cannot join elements in path err: %w", err)
	}

	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	req, err := c.prepareRequest(rctx, body, url)
	if err != nil {
		return fmt.Errorf("cannot prepare request err: %w", err)
	}

	return c.DoRequest(req)

}

func (c *Client) BatchUpdate(ctx context.Context, metrics []*metrics.Metrics) error {

	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("cannot marshal metric err: %w", err)
	}

	url, err := url.JoinPath("http://", c.host, "/updates/")
	if err != nil {
		return fmt.Errorf("cannot join elements in path err: %w", err)
	}

	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	req, err := c.prepareRequest(rctx, body, url)
	if err != nil {
		return fmt.Errorf("cannot prepare request err: %w", err)
	}

	return c.DoRequest(req)

}

func (c *Client) DoRequest(req *retryablehttp.Request) error {

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request execute err: %w", err)
	}

	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body err: %w", err)
	}

	if resp.StatusCode < 300 {
		c.logger.Log.Infof(`request for update metric has been completed
	code: %d`, resp.StatusCode)
	} else {
		c.logger.Log.Errorf(`request for update metric has failed
	code: %d
	result: %s`, resp.StatusCode, string(res))
	}

	return nil

}
