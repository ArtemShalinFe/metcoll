// Package metcoll for interacting with the metrics server.
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

// Client - sends requests for metric updates to the server.
type Client struct {
	host       string
	httpClient *retryablehttp.Client
	logger     retryablehttp.LeveledLogger
	hashkey    []byte
}

// NewClient - Object constructor.
func NewClient(cfg *configuration.ConfigAgent, logger retryablehttp.LeveledLogger) *Client {
	const defautMaxRetry = 3
	const defautMinWaitRetry = 3
	const defautMaxWaitRetry = 5

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = defautMaxRetry
	retryClient.CheckRetry = checkRetry
	retryClient.RetryWaitMin = time.Duration(defautMinWaitRetry * time.Second)
	retryClient.RetryWaitMax = time.Duration(defautMaxWaitRetry * time.Second)
	retryClient.Logger = logger
	retryClient.Backoff = backoff

	return &Client{
		host:       cfg.Server,
		httpClient: retryClient,
		logger:     logger,
		hashkey:    cfg.Key,
	}
}

func checkRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true, err
	} else {
		return false, err
	}
}

func backoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	const an0 = 0
	const an1 = 1
	const an2 = 2

	const an0backoff = 1 * time.Second
	const an1backoff = 3 * time.Second
	const an2backoff = 5 * time.Second
	const defaultbackoff = 2 * time.Second

	switch attemptNum {
	case an0:
		return an0backoff
	case an1:
		return an1backoff
	case an2:
		return an2backoff
	default:
		return defaultbackoff
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
	if len(c.hashkey) != 0 {
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

	if resp.StatusCode < http.StatusMultipleChoices {
		c.logger.Info(`request for update metric has been completed
	code: %d, hash: %s`, resp.StatusCode, resp.Header.Get("HashSha256"))
	} else {
		c.logger.Error(`request for update metric has failed
	code: %d
	result: %s`, resp.StatusCode, string(res))
	}

	return nil
}

// PushResult - Consolidates the result of sending the metric and the error.
type PushResult struct {
	Metric *metrics.Metrics
	Err    error
}

// BatchUpdateMetric - Sends updated metrics received from the channel `mcs` to the server.
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
