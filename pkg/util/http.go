package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpInterface interface {
	Post(url string, val interface{}, data interface{}, header map[string]string) error
	Put(url string, val interface{}, data map[string]interface{}) error
	Delete(url string, val interface{}) error
	Get(url string, val interface{}) error
}

type Auth struct {
	Username string
	Password string
}

type httpClient struct {
	timeout time.Duration
	url     string
	auth    *Auth
	headers map[string]string
}

func NewHttpClient(timeout time.Duration, url string) *httpClient {
	return &httpClient{timeout: timeout, url: url}
}

func (c *httpClient) WithAuth(username, password string) {
	c.auth = &Auth{
		Username: username,
		Password: password,
	}
}

func (c *httpClient) WithHeaders(headers map[string]string) {
	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	// 追加请求头
	for k, v := range headers {
		c.headers[k] = v
	}
}

func (c *httpClient) isSuccess(statusCode int) bool {
	return statusCode == http.StatusOK
}

func (c *httpClient) parse(r io.Reader, val interface{}) error {
	if val != nil {
		d, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(d, val); err != nil {
			return err
		}
	}
	return nil
}

func (c *httpClient) Get(url string, val interface{}) error {
	client := &http.Client{Timeout: c.timeout}
	req, err := http.NewRequest("", url, nil)
	if err != nil {
		return err
	}
	if c.auth != nil {
		req.SetBasicAuth(c.auth.Username, c.auth.Password)
	}
	if c.headers != nil {
		for key, value := range c.headers {
			req.Header.Set(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !c.isSuccess(resp.StatusCode) {
		return fmt.Errorf("error resp %s", resp.Status)
	}
	return c.parse(resp.Body, val)
}

func (c *httpClient) Post(url string, val interface{}, data interface{}, header map[string]string) error {
	client := &http.Client{Timeout: c.timeout}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// 设置请求头
	if header != nil {
		for key, value := range header {
			req.Header.Set(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error resp %s", resp.Status)
	}

	if val != nil {
		d, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(d, val); err != nil {
			return err
		}
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

	if val != nil {
		d, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(d, val); err != nil {
			return err
		}
	}

	return nil
}

func (c *httpClient) Delete(url string, val interface{}) error {
	client := &http.Client{Timeout: c.timeout}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	if c.auth != nil {
		req.SetBasicAuth(c.auth.Username, c.auth.Password)
	}
	if c.headers != nil {
		for key, value := range c.headers {
			req.Header.Set(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !c.isSuccess(resp.StatusCode) {
		return fmt.Errorf("error resp %s", resp.Status)
	}
	return c.parse(resp.Body, val)
}
