package client

import (
	"distributed_cache/common"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Client struct {
	serverAddr string
}

func NewClient(addr string) *Client {
	return &Client{serverAddr: addr}
}

func (c *Client) log(format string, v ...any) {
	if common.DEBUG {
		log.Printf(format, v...)
	}
}

func (c *Client) ServerAddr() string {
	return c.serverAddr
}

func (c *Client) Get(serviceName string, key string) ([]byte, error) {
	c.log("Client [GET]: service[%s] key[%s]", serviceName, key)
	url := fmt.Sprintf("%v%v/%v", c.serverAddr, serviceName, key)
	resp, err := http.Get(url)
	if err != nil {
		c.log("request from %s error %s", url, err)
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		bytes, err := io.ReadAll(resp.Body)
		return bytes, err
	default:
		c.log("Client [ERROR] response status: %s", resp.Status)
		return nil, errors.New(resp.Status)
	}
}
