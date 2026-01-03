package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HttpClientV2 struct {
	URL    string
	method string

	auth    *Auth
	body    io.Reader
	headers map[string]string
	timeout time.Duration
}

func (c *HttpClientV2) Method(method string) *HttpClientV2 {
	if c == nil {
		return nil
	}
	c.method = method
	return c
}

func (c *HttpClientV2) WithTimeout(t time.Duration) *HttpClientV2 {
	if c == nil {
		return nil
	}
	c.timeout = t
	return c
}

func (c *HttpClientV2) WithAuth(username, password string) *HttpClientV2 {
	if c == nil {
		return nil
	}
	c.auth = &Auth{Username: username, Password: password}
	return c
}

func (c *HttpClientV2) WithHeader(headers map[string]string) *HttpClientV2 {
	if c == nil {
		return nil
	}
	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	// 追加请求头
	for k, v := range headers {
		c.headers[k] = v
	}
	return c
}

func (c *HttpClientV2) Do(val interface{}) error {
	if c == nil {
		return fmt.Errorf("httpClient is nil")
	}

	req, err := http.NewRequest(c.method, c.URL, c.body)
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

	client := &http.Client{Timeout: c.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error resp %s", resp.Status)
	}

	if val != nil {
		d, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(d, val); err != nil {
			return err
		}
	}
	return nil
}
