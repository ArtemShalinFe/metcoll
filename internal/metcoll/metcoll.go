package metcoll

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Client struct {
	host string
}

func NewClient(Host string) *Client {
	return &Client{
		host: Host,
	}
}

func (c *Client) Update(j json.Marshaler) error {

	b, err := json.Marshal(j)
	if err != nil {
		return err
	}

	body := bytes.NewBuffer(b)
	resp, err := http.Post(fmt.Sprintf("http://%s/update/", c.host), "application/json", body)
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
