package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type HttpError struct {
	Message string
	Code    int
	Err     error
}

func (h HttpError) Error() string {
	if h.Err != nil {
		return fmt.Sprintf("%d %v", h.Code, h.Err)
	}
	return h.Message
}

type HttpClientV2 struct {
	URL    string
	method string

	filename *string
	auth     *Auth
	body     io.Reader
	headers  map[string]string
	timeout  time.Duration
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

func (c *HttpClientV2) WithBody(body io.Reader) *HttpClientV2 {
	if c == nil {
		return nil
	}
	c.body = body
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

func (c *HttpClientV2) WithFile(filename string) *HttpClientV2 {
	if c == nil {
		return nil
	}
	c.filename = &filename
	return c
}

func (c *HttpClientV2) Do(val interface{}) *HttpError {
	if c == nil {
		return &HttpError{
			Message: "httpClient is nil",
		}
	}

	req, err := http.NewRequest(c.method, c.URL, c.body)
	if err != nil {
		return &HttpError{
			Err: err,
		}
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
		return &HttpError{
			Err: err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		d, err := io.ReadAll(resp.Body)
		if err != nil {
			return &HttpError{Err: err}
		}
		return &HttpError{Message: string(d), Code: resp.StatusCode}
	}

	// 结果存入文件
	if c.filename != nil {
		file, err := os.Create(*c.filename)
		if err != nil {
			return &HttpError{
				Err: fmt.Errorf("创建文件失败: %w", err),
			}
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return &HttpError{
				Err: fmt.Errorf("写入文件失败: %w", err),
			}
		}
	}

	// 结果存入结构体
	if val != nil {
		d, err := io.ReadAll(resp.Body)
		if err != nil {
			return &HttpError{
				Err: err,
			}
		}
		if err = json.Unmarshal(d, val); err != nil {
			return &HttpError{
				Err: err,
			}
		}
	}
	return nil
}
