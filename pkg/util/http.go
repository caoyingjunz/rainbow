package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpInterface interface {
	Post(url string, val interface{}, data map[string]interface{}) error
	Put(url string, val interface{}, data map[string]interface{}) error
	Get(url string, val interface{}) error
}

type httpClient struct {
	timeout time.Duration
	url     string
}

func NewHttpClient(timeout time.Duration, url string) *httpClient {
	return &httpClient{timeout: timeout, url: url}
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error resp %s", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, val); err != nil {
		return err
	}

	return nil
}

func (c *httpClient) Post(url string, val interface{}, data map[string]interface{}) error {
	client := &http.Client{Timeout: c.timeout}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error resp %s", resp.Status)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(d, val); err != nil {
		return err
	}

	return nil
}

func (c *httpClient) Put(url string, val interface{}, data map[string]interface{}) error {
	client := &http.Client{Timeout: c.timeout}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error resp %s", resp.Status)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(d, val); err != nil {
		return err
	}

	return nil
}
