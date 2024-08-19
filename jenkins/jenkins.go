package jenkins

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"strings"
)

const (
	crumbURL        = "/crumbIssuer/api/json"
	defaultBaseURL  = "http://127.0.0.1:8080"
	defaultUserName = "admin"
)

type Crumbs struct {
	Value        string `json:"crumb"`
	RequestField string `json:"crumbRequestField"`
}

type BasicAuthTransport struct {
	Username string
	Password string
}

func (t BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.Username, t.Password)
	return http.DefaultTransport.RoundTrip(req)
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	userName   string
	password   string
	apiToken   string
	userAgent  string
	crumbs     *Crumbs

	common service
	Nodes  *NodesService
}

type service struct {
	client *Client
}

type ClientOption func(*Client) error

// WithClient sets the http client for the Jenkins client
func WithClient(client *http.Client) ClientOption {
	return func(c *Client) error {
		c.httpClient = client
		return nil
	}
}

// WithBaseURL sets the base URL for the Jenkins client
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		c.baseURL = baseURL
		return nil
	}
}

// WithUserPassword sets the password for the Jenkins client
func WithUserPassword(userName, password string) ClientOption {
	return func(c *Client) error {
		if c.apiToken != "" {
			return fmt.Errorf("cannot set both API token and password")
		}
		c.userName = userName
		c.password = password
		return nil
	}
}

// WithUserToken sets the API token for the Jenkins client
func WithUserToken(userName, apiToken string) ClientOption {
	return func(c *Client) error {
		if c.password != "" {
			return fmt.Errorf("cannot set both API token and password")
		}
		c.userName = userName
		c.apiToken = apiToken
		return nil
	}
}

// NewClient returns a new Jenkins API client
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		baseURL:  defaultBaseURL,
		userName: defaultUserName,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.httpClient == nil {
		jar, _ := cookiejar.New(nil)
		c.httpClient = &http.Client{Jar: jar}
	}

	if c.apiToken != "" || c.password != "" {
		c.httpClient.Transport = &BasicAuthTransport{
			Username: c.userName,
			Password: c.apiToken,
		}
		if c.password != "" {
			c.httpClient.Transport.(*BasicAuthTransport).Password = c.password
		}
	}

	c.common.client = c
	c.Nodes = (*NodesService)(&c.common)

	return c, nil
}

func (c *Client) setCrumbs(ctx context.Context) error {
	resp, err := c.get(ctx, crumbURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var crumbs Crumbs
	if err := json.NewDecoder(resp.Body).Decode(&crumbs); err != nil {
		return err
	}

	c.crumbs = &crumbs
	return nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	u, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return req, nil
}

func (c *Client) newFormRequest(ctx context.Context, path string, values url.Values) (*http.Request, error) {
	req, err := c.newRequest(ctx, http.MethodPost, path, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if c.crumbs != nil {
		req.Header.Add(c.crumbs.RequestField, c.crumbs.Value)
		c.crumbs = nil
	}

	return req, nil
}

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return resp, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	return resp, nil
}

func convertBodyStruct(body interface{}) url.Values {
	values := make(url.Values)
	v := reflect.ValueOf(body).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" {
			tag = field.Name
		}
		values.Set(tag, fmt.Sprint(v.Field(i)))
	}

	return values
}

func (c *Client) postForm(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	if err := c.setCrumbs(ctx); err != nil {
		return nil, err
	}

	values := convertBodyStruct(body)

	req, err := c.newFormRequest(ctx, path, values)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return resp, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return resp, nil
}

func (c *Client) post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	if err := c.setCrumbs(ctx); err != nil {
		return nil, err
	}

	b, err := xml.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodPost, path, strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/xml")

	if c.crumbs != nil {
		req.Header.Add(c.crumbs.RequestField, c.crumbs.Value)
		c.crumbs = nil
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return resp, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return resp, nil
}
