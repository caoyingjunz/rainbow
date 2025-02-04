package util

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type httpClient struct {
	timeout time.Duration
}

func NewHttpClient(timeout time.Duration) *httpClient {
	return &httpClient{timeout}
}

func (c *httpClient) Get(url string, val interface{}) error {
	client := &http.Client{Timeout: c.timeout}
	req, err := http.NewRequest("", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, val); err != nil {
		return err
	}

	return nil
}
