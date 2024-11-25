package utils

import (
	"fadacontrol/pkg/secure"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/net/proxy"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
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
func (c *Client) PostWithSign(url, accessKey, accessSecret string, signAlgorithm secure.SignAlgorithm, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("x-rcs-key", accessKey)
	req.Header.Set("x-rcs-timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	if req.Header.Get("x-rcs-nonce") == "" {
		req.Header.Set("x-rcs-nonce", uuid.New().String())
	}
	req.Header.Set("x-rcs-signature-method", string(signAlgorithm))
	signatureString, err := generateSignatureString(req)
	if err != nil {
		return nil, err
	}
	signature, err := secure.CalculateHMAC(signatureString, accessSecret, signAlgorithm)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-rcs-signature", signature)
	return c.httpClient.Do(req)

}
func (c *Client) GetWithSign(url, accessKey, accessSecret string, signAlgorithm secure.SignAlgorithm, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("x-rcs-key", accessKey)
	req.Header.Set("x-rcs-timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	if req.Header.Get("x-rcs-nonce") == "" {
		req.Header.Set("x-rcs-nonce", uuid.New().String())
	}

	req.Header.Set("x-rcs-signature-method", string(signAlgorithm))
	signatureString, err := generateSignatureString(req)
	if err != nil {
		return nil, err
	}
	signature, err := secure.CalculateHMAC(signatureString, accessSecret, signAlgorithm)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-rcs-signature", signature)
	return c.httpClient.Do(req)

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
func generateSignatureString(req *http.Request) (string, error) {
	// Extract HTTP Method
	httpMethod := strings.ToUpper(req.Method)

	// Extract Accept
	accept := req.Header.Get("Accept")

	// Extract Content-MD5
	contentMD5 := req.Header.Get("Content-MD5")

	// Extract Content-Type
	contentType := req.Header.Get("Content-Type")

	// Extract Date
	date := req.Header.Get("Date")

	// Extract Headers
	headers, headerKeys, err := extractHeaders(req)
	if err != nil {
		return "", err
	}
	// Set x-rcs-signature-headers
	if len(headerKeys) > 0 {
		req.Header.Set("x-rcs-signature-headers", strings.Join(headerKeys, ","))
	}

	// Extract Path and Parameters
	pathAndParameters := extractPathAndParameters(req)

	// Construct Signature String
	signatureString := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s%s",
		httpMethod,
		accept,
		contentMD5,
		contentType,
		date,
		headers,
		pathAndParameters,
	)

	return signatureString, nil
}

func extractHeaders(req *http.Request) (string, []string, error) {
	var headers []string
	var headerKeys []string
	for key, values := range req.Header {
		if strings.HasPrefix(strings.ToLower(key), "x-rcs-") && strings.ToLower(key) != "x-rcs-signature" && strings.ToLower(key) != "x-rcs-signature-headers" {
			headerKeys = append(headerKeys, strings.ToLower(key))
			headerValue := ""
			if len(values) > 0 {
				headerValue = values[0]
			}
			headers = append(headers, fmt.Sprintf("%s:%s\n", strings.ToLower(key), headerValue))
		}
	}

	sort.Strings(headers)
	sort.Strings(headerKeys)
	return strings.Join(headers, ""), headerKeys, nil
}
func extractPathAndParameters(req *http.Request) string {
	path := req.URL.Path

	queryParams := req.URL.Query()
	keys := make([]string, 0, len(queryParams))
	for key := range queryParams {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var queryParts []string
	for _, key := range keys {
		values := queryParams[key]
		if len(values) > 0 {
			queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, values[0]))
		} else {
			queryParts = append(queryParts, key)
		}
	}

	if len(queryParts) > 0 {
		return fmt.Sprintf("%s?%s", path, strings.Join(queryParts, "&"))
	}

	return path
}
