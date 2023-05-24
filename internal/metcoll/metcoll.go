package metcoll

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/ArtemShalinFe/metcoll/internal/metrics"
)

type Logger interface {
	Info(args ...interface{})
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

func (c *Client) prepareRequest(metric json.Marshaler) (*http.Request, error) {

	body, err := json.Marshal(metric)
	if err != nil {
		return nil, err
	}

	var zBuf bytes.Buffer
	zw := gzip.NewWriter(&zBuf)

	if _, err := zw.Write(body); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	url, err := url.JoinPath("http://", c.host, "/update/")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, &zBuf)
	if err != nil {
		return nil, err
	}
	req.Close = true

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	return req, nil
}

func (c *Client) Update(metric *metrics.Metrics) error {

	req, err := c.prepareRequest(metric)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	c.logger.Info("request for update metric has been completed ",
		"code: ", resp.StatusCode,
		"result: ", string(res))

	return nil
}
