package metcoll

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"syscall"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
)

type Client struct {
	host       string
	httpClient *retryablehttp.Client
	logger     retryablehttp.LeveledLogger
	hashkey    string
}

func NewClient(cfg *configuration.ConfigAgent, logger retryablehttp.LeveledLogger) *Client {

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.CheckRetry = CheckRetry
	retryClient.RetryWaitMin = time.Duration(3 * time.Second)
	retryClient.RetryWaitMax = time.Duration(5 * time.Second)
	retryClient.Logger = logger
	retryClient.Backoff = Backoff

	return &Client{
		host:       cfg.Server,
		httpClient: retryClient,
		logger:     logger,
		hashkey:    cfg.Key,
	}

}

func CheckRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {

	if errors.Is(err, syscall.ECONNREFUSED) {
		return true, err
	} else {
		return false, err
	}

}

func Backoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {

	switch attemptNum {
	case 0:
		return 1 * time.Second
	case 1:
		return 3 * time.Second
	case 2:
		return 5 * time.Second
	default:
		return 2 * time.Second
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
	if c.hashkey != "" {

		data, err := req.BodyBytes()
		if err != nil {
			return nil, fmt.Errorf("cannot calculate hash err: %w", err)
		}

		h := hmac.New(sha256.New, []byte(c.hashkey))

		h.Write(data)

		req.Header.Set("HashSHA256", fmt.Sprintf("%x", h.Sum(nil)))

	}

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

	req, err := c.prepareRequest(ctx, body, url)
	if err != nil {
		return fmt.Errorf("cannot prepare request err: %w", err)
	}

	return c.doRequest(req)

}

func (c *Client) batchUpdate(ctx context.Context, metrics []*metrics.Metrics) error {

	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("cannot marshal metric err: %w", err)
	}

	url, err := url.JoinPath("http://", c.host, "/updates/")
	if err != nil {
		return fmt.Errorf("cannot join elements in path err: %w", err)
	}

	req, err := c.prepareRequest(ctx, body, url)
	if err != nil {
		return fmt.Errorf("cannot prepare request err: %w", err)
	}

	return c.doRequest(req)

}

func (c *Client) doRequest(req *retryablehttp.Request) error {

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
		c.logger.Info(`request for update metric has been completed
	code: %d`, resp.StatusCode)
	} else {
		c.logger.Error(`request for update metric has failed
	code: %d
	result: %s`, resp.StatusCode, string(res))
	}

	return nil

}

type PushResult struct {
	Metric *metrics.Metrics
	Err    error
}

func (c *Client) UpdateMetric(ctx context.Context, ms <-chan *metrics.Metrics, results chan<- PushResult) {

	for m := range ms {

		pr := &PushResult{
			Metric: m,
			Err:    nil,
		}

		if err := c.Update(ctx, m); err != nil {
			pr.Err = fmt.Errorf("push metric %s on server err: %w", m.ID, err)
		}

		select {
		case <-ctx.Done():
			return
		case results <- *pr:
		}

	}

}

func (c *Client) BatchUpdateMetric(ctx context.Context, mcs <-chan []*metrics.Metrics, result chan<- error) {

	for m := range mcs {

		err := c.batchUpdate(ctx, m)

		select {
		case <-ctx.Done():
			return
		case result <- err:
		}

	}

}
