package client

import (
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

func (c *Client) Push(mType string, Name string, Value string) {

	resp, err := http.Post(fmt.Sprintf("http://%s/update/%s/%s/%s", c.host, mType, Name, Value), "text/plain", nil)
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer resp.Body.Close()
	userResult, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err.Error())
	}
	log.Printf("Resp: [%d] [%s]\n", resp.StatusCode, string(userResult))
}
