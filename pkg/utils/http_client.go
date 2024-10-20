package utils

import (
	"golang.org/x/net/proxy"
	"io"
	"net/http"
	"net/url"
	"time"
)

type ClientBuilder struct {
	proxyURL string
	timeout  time.Duration
}

func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{}
}
func (b *ClientBuilder) SetProxy(proxyURL string) *ClientBuilder {
	b.proxyURL = proxyURL
	return b
}

func (b *ClientBuilder) SetTimeout(timeout time.Duration) *ClientBuilder {
	b.timeout = timeout
	return b
}
func (b *ClientBuilder) Build() (*Client, error) {
	client := &Client{

		httpClient: &http.Client{
			Timeout: b.timeout,
		},
	}

	if b.proxyURL != "" {
		err := client.SetProxy(b.proxyURL)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

// Client
type Client struct {
	httpClient *http.Client
}

// NewClient
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

// SetProxy
func (c *Client) SetProxy(proxyURL string) error {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return err
	}

	switch parsedURL.Scheme {
	case "socks5":
		dialer, err := proxy.SOCKS5("tcp", parsedURL.Host, nil, proxy.Direct)
		if err != nil {
			return err
		}
		transport := &http.Transport{
			Dial: dialer.Dial,
		}
		c.httpClient.Transport = transport
	case "http", "https":
		transport := &http.Transport{
			Proxy: http.ProxyURL(parsedURL),
		}
		c.httpClient.Transport = transport
	default:
		return nil
	}

	return nil
}

// Get
func (c *Client) Get(url string) (string, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Post
func (c *Client) Post(url string, contentType string, body io.Reader, headers map[string]string) (string, error) {

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", contentType)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}
