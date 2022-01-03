// Copyright 2021 The go-jenkins AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
	// crumbURL is the URL to issue a crumb request.
	crumbURL       = "/crumbIssuer/api/json"
	defaultBaseURL = "http://127.0.0.1:8080"
)

// Crumbs represents Jenkins CSRF Crumbs
type Crumbs struct {
	Value        string `json:"crumb"`
	RequestField string `json:"crumbRequestField"`
}

type BasicAuthTransport struct {
	Username string
	Password string
}

func (bat BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(bat.Username, bat.Password)

	return http.DefaultTransport.RoundTrip(req)
}

// A Client manages communication with the Jenkins API.
type Client struct {
	httpClient *http.Client

	UserAgent string

	Crumbs *Crumbs

	common service

	baseURL  string
	userName string
	password string
	apiToken string

	Nodes *NodesService
}

type service struct {
	client *Client
}

type ClientOption func(*Client) error

func WithClient(client *http.Client) ClientOption {
	return func(c *Client) error {
		c.httpClient = client

		return nil
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		c.baseURL = baseURL

		return nil
	}
}

func WithPassword(userName, password string) ClientOption {
	return func(c *Client) error {
		if c.apiToken != "" {
			return fmt.Errorf("cannot set both API token and password")
		}
		c.userName = userName
		c.password = password

		return nil
	}
}

func WithToken(userName, apiToken string) ClientOption {
	return func(c *Client) error {
		if c.password != "" {
			return fmt.Errorf("cannot set both API token and password")
		}
		c.userName = userName
		c.apiToken = apiToken

		return nil
	}
}

func DefaultHTTPClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar}
}

// NewClient returns a new Jenkins API client.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{baseURL: defaultBaseURL}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.httpClient == nil {
		c.httpClient = DefaultHTTPClient()
	}

	if c.apiToken != "" {
		c.httpClient.Transport = BasicAuthTransport{Username: c.userName, Password: c.apiToken}
	}

	if c.password != "" {
		c.httpClient.Transport = BasicAuthTransport{Username: c.userName, Password: c.password}
	}

	c.common.client = c
	c.Nodes = (*NodesService)(&c.common)

	return c, nil
}

// setCrumbs sets the Crumbs for the client.
func (c *Client) setCrumbs(ctx context.Context) (*http.Response, error) {
	resp, err := c.get(ctx, crumbURL)
	if err != nil {
		return resp, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}

	crumbs := &Crumbs{}
	err = json.Unmarshal(body, crumbs)
	if err != nil {
		return resp, err
	}

	c.Crumbs = crumbs

	return resp, nil
}

// newRequest creates an API request. A relative URL can be provided in query,
func (c *Client) newRequest(ctx context.Context, method string, query string, body io.Reader) (*http.Request, error) {
	query = "/" + strings.TrimPrefix(query, "/")
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("%s%s", c.baseURL, query),
		body,
	)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// newFormRequest creates an API request with form data.
func (c *Client) newFormRequest(ctx context.Context, query string, values url.Values) (*http.Request, error) {
	body := strings.NewReader(values.Encode())
	req, err := c.newRequest(ctx, "POST", query, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if c.Crumbs != nil {
		req.Header.Add(c.Crumbs.RequestField, c.Crumbs.Value)
		// Crumbs are only valid for one request.
		c.Crumbs = nil
	}
	return req, nil
}

// get issues a GET to the specified path.
func (c *Client) get(ctx context.Context, query string) (*http.Response, error) {
	req, err := c.newRequest(ctx, "GET", query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode > 299 {
		return resp, fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
	}

	if c.httpClient.Jar != nil {
		for _, cookie := range resp.Cookies() {
			c.httpClient.Jar.SetCookies(req.URL, []*http.Cookie{cookie})
		}
	}

	return resp, nil
}

// convertBodyStruct is a helper function to convert a struct to url.Values.
func convertBodyStruct(body interface{}) url.Values {
	values := make(url.Values)
	iVal := reflect.ValueOf(body).Elem()
	typ := iVal.Type()
	for i := 0; i < iVal.NumField(); i++ {
		j := typ.Field(i).Tag.Get("json")
		if j == "" {
			j = typ.Field(i).Name
		}
		values.Set(j, fmt.Sprint(iVal.Field(i)))
	}

	return values
}

// postForm issues a POST to the specified path with the given form data.
func (c *Client) postForm(ctx context.Context, query string, body interface{}) (*http.Response, error) {
	crumbsResp, err := c.setCrumbs(ctx)
	if err != nil {
		return crumbsResp, err
	}

	values := convertBodyStruct(body)

	req, _ := c.newFormRequest(ctx, query, values)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return resp, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode > 299 {
		return resp, fmt.Errorf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	if c.httpClient.Jar != nil {
		for _, cookie := range resp.Cookies() {
			c.httpClient.Jar.SetCookies(req.URL, []*http.Cookie{cookie})
		}
	}

	return resp, nil
}

func (c *Client) post(ctx context.Context, query string, body interface{}) (*http.Response, error) {
	crumbsResp, err := c.setCrumbs(ctx)
	if err != nil {
		return crumbsResp, err
	}

	j, err := xml.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyR := strings.NewReader(string(j))
	req, _ := c.newRequest(ctx, "POST", query, bodyR)
	req.Header.Set("Content-Type", "application/xml")

	if c.Crumbs != nil {
		req.Header.Add(c.Crumbs.RequestField, c.Crumbs.Value)
		// Crumbs are only valid for one request.
		c.Crumbs = nil
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode > 299 {
		return resp, fmt.Errorf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	if c.httpClient.Jar != nil {
		for _, cookie := range resp.Cookies() {
			c.httpClient.Jar.SetCookies(req.URL, []*http.Cookie{cookie})
		}
	}

	return resp, nil
}
