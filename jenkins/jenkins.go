// Copyright 2021 The go-jenkins AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jenkins

import (
	"context"
	"encoding/json"
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
	crumbURL = "/crumbIssuer/api/json"
)

// Crumbs represents Jenkins CSRF Crumbs
type Crumbs struct {
	Value        string `json:"crumb"`
	RequestField string `json:"crumbRequestField"`
}

// A Client manages communication with the Jenkins API.
type Client struct {
	httpClient *http.Client

	UserAgent string

	Crumbs *Crumbs

	common service

	baseURL  string
	userName string
	apiToken string

	Nodes *NodesService
}

type service struct {
	client *Client
}

// NewClient returns a new Jenkins API client.
func NewClient(httpClient *http.Client, baseURL string, userName string, apiToken string) *Client {
	if httpClient == nil {
		jar, _ := cookiejar.New(nil)
		httpClient = &http.Client{Jar: jar}
	}

	c := &Client{httpClient: httpClient, baseURL: baseURL, userName: userName, apiToken: apiToken}
	c.common.client = c
	c.Nodes = (*NodesService)(&c.common)

	return c
}

// SetCrumbs sets the Crumbs for the client.
func (c *Client) SetCrumbs(ctx context.Context) (*http.Response, error) {
	resp, err := c.Get(ctx, crumbURL)
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

// NewRequest creates an API request. A relative URL can be provided in query,
func (c *Client) NewRequest(ctx context.Context, method string, query string, body io.Reader) (*http.Request, error) {
	query = strings.TrimPrefix(query, "/")
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("%s%s", c.baseURL, query),
		body,
	)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.userName, c.apiToken)

	return req, nil
}

// NewFormRequest creates an API request with form data.
func (c *Client) NewFormRequest(ctx context.Context, query string, body io.Reader) (*http.Request, error) {
	req, err := c.NewRequest(ctx, "POST", query, body)
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

// Get issues a GET to the specified path.
func (c *Client) Get(ctx context.Context, query string) (*http.Response, error) {
	req, err := c.NewRequest(ctx, "GET", query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return resp, err
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

// PostForm issues a POST to the specified path with the given form data.
func (c *Client) PostForm(ctx context.Context, query string, body interface{}) (*http.Response, error) {
	crumbsResp, err := c.SetCrumbs(ctx)
	if err != nil {
		return crumbsResp, err
	}

	values := convertBodyStruct(body)

	req, _ := c.NewFormRequest(ctx, query, strings.NewReader(values.Encode()))

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
