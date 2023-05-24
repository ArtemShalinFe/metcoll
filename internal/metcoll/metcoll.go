package metcoll

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ArtemShalinFe/metcoll/internal/metrics"
)

type Client struct {
	host       string
	httpClient *http.Client
}

func NewClient(Host string) *Client {

	return &Client{
		host:       Host,
		httpClient: &http.Client{},
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

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/update/", c.host), &zBuf)
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

	log.Printf("Resp: [%d] [%s]\n", resp.StatusCode, string(res))

	return nil
}
