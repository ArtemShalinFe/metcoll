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

	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/sleepstepper"
)

type Logger interface {
	Infof(template string, args ...interface{})
	Errorf(template string, args ...interface{})
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

func (c *Client) prepareRequest(ctx context.Context, body []byte, url string) (*http.Request, error) {

	var zBuf bytes.Buffer
	zw := gzip.NewWriter(&zBuf)

	if _, err := zw.Write(body); err != nil {
		return nil, fmt.Errorf("cannot write compressed body err: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("cannot close compress writer err: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &zBuf)
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

func (c *Client) DoRequest(req *http.Request) error {

	ss := sleepstepper.NewSleepStepper(1, 2, 5)
	resp, err := retryHTTPrequest(c.httpClient.Do, req, ss)
	if err != nil {
		return fmt.Errorf("request execute err: %w", err)
	}

	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body err: %w", err)
	}

	if resp.StatusCode < 300 {
		c.logger.Infof(`request for update metric has been completed
	code: %d`, resp.StatusCode)
	} else {
		c.logger.Errorf(`request for update metric has failed
	code: %d
	result: %s`, resp.StatusCode, string(res))
	}

	return nil

}

type Func func(req *http.Request) (*http.Response, error)

type Sleeper interface {
	Sleep() bool
}

func retryHTTPrequest(f Func, req *http.Request, ss Sleeper) (*http.Response, error) {

	res, err := f(req)

	if err != nil {

		if !errors.Is(err, syscall.ECONNREFUSED) {
			return nil, err
		}

		if !ss.Sleep() {
			return nil, err
		}

		return retryHTTPrequest(f, req, ss)

	}

	return res, nil

}
