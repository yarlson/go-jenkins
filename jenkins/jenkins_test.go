// Copyright 2021 The go-jenkins AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jenkins

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type Suite struct {
	mux    *http.ServeMux
	server *httptest.Server

	suite.Suite
}

func (s *Suite) newMux() {
	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)
}

func (s *Suite) addCrumbsHandle() {
	s.mux.HandleFunc(crumbURL, func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`{"crumbRequestField":"crumb", "crumb":"crumb"}`,
		))
		s.NoError(err)
	})
}

func TestSuite(t *testing.T) {
	s := new(Suite)

	suite.Run(t, s)
}

func (s *Suite) TestConvertBodyStruct() {
	type testBody struct {
		Name string `json:"name"`
	}

	s.Equal(convertBodyStruct(&testBody{Name: "test"}), url.Values{"name": []string{"test"}})
}

func (s *Suite) TestConvertBodyStructNoJsonTag() {
	type testBody struct {
		Name string
	}

	s.Equal(convertBodyStruct(&testBody{Name: "test"}), url.Values{"Name": []string{"test"}})
}

func (s *Suite) TestNewClient() {
	_, err := NewClient()
	s.NoError(err)
}

func (s *Suite) TestNewClientWithClient() {
	_, err := NewClient(WithClient(&http.Client{}))
	s.NoError(err)
}

func (s *Suite) TestNewClientWithPassword() {
	_, err := NewClient(WithUserPassword("test", "test"))
	s.NoError(err)
}

func (s *Suite) TestNewClientWithToken() {
	_, err := NewClient(WithUserToken("test", "test"))
	s.NoError(err)
}

func (s *Suite) TestNewClientWithTokenAndPassword() {
	_, err := NewClient(WithUserToken("test", "test"), WithUserPassword("test", "test"))
	s.Error(err)
}

func (s *Suite) TestNewClientWithPasswordAndToken() {
	_, err := NewClient(WithUserPassword("test", "test"), WithUserToken("test", "test"))
	s.Error(err)
}

func (s *Suite) TestClientNewRequest() {
	client, err := NewClient()
	s.NoError(err)

	_, err = client.newRequest(context.Background(), "GET", "/", nil)

	s.NoError(err)
}

func (s *Suite) TestClientNewRequestError() {
	client, err := NewClient()
	s.NoError(err)

	//lint:ignore SA1012 this is a test
	//nolint
	_, err = client.newRequest(nil, "GET", "/", nil)
	s.Error(err)
}

func (s *Suite) testMethod(r *http.Request, want string) {
	s.Equal(want, r.Method)
}

func (s *Suite) TestClientGet() {
	s.newMux()
	s.mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`{"A":"a"}`,
		))
		s.NoErrorf(err, "w.Write returned %v")
		s.Equal("Basic YWRtaW46YWRtaW4=", r.Header.Get("Authorization"))
	})

	client, err := NewClient(WithBaseURL(s.server.URL), WithUserPassword("admin", "admin"))
	s.NoError(err)

	got, err := client.get(context.Background(), "test")
	s.NoError(err)
	s.Equal(got.StatusCode, http.StatusOK)

	all, err := io.ReadAll(got.Body)
	s.NoError(err)
	s.Equal(string(all), `{"A":"a"}`)
}

func (s *Suite) TestClientGetDoError() {
	s.newMux()
	s.mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		// Sends zero body with wrong content-length
		w.Header().Set("Content-Length", "1")
	})

	client, err := NewClient(WithBaseURL(s.server.URL), WithUserPassword("admin", "admin"))
	s.NoError(err)

	_, err = client.get(context.Background(), "test")
	s.NoError(err)
}

func (s *Suite) TestClientGetNotFound() {
	s.newMux()

	client, err := NewClient(WithBaseURL(s.server.URL))
	s.NoError(err)

	got, err := client.get(context.Background(), "test_error")
	s.Error(err)
	s.Equal(got.StatusCode, http.StatusNotFound)
}

func (s *Suite) TestClientGetErrorContext() {
	client, err := NewClient()
	s.NoError(err)

	//lint:ignore SA1012 this is a test
	//nolint
	_, err = client.get(nil, "test_error")
	s.Error(err)
}

func (s *Suite) TestClientGetErrorResponse() {
	client, err := NewClient()
	s.NoError(err)

	deadCtx, cancel := context.WithDeadline(context.Background(), time.Now())
	defer cancel()

	_, err = client.get(deadCtx, "test_error")
	s.Error(err)
}

func (s *Suite) TestClientGetCookie() {
	s.newMux()
	s.mux.HandleFunc("/test_cookie", func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		w.Header().Set("Set-Cookie", "test=cookie")
	})

	client, err := NewClient(WithBaseURL(s.server.URL), WithUserPassword("admin", "admin"))
	s.NoError(err)

	got, err := client.get(context.Background(), "test_cookie")
	s.NoError(err)
	s.Equal(got.StatusCode, http.StatusOK)

	_, err = io.ReadAll(got.Body)
	s.NoError(err)
	s.Equal(got.Cookies()[0].Name, "test")
	s.Equal(got.Cookies()[0].Value, "cookie")
}

func (s *Suite) TestClientNewFormRequest() {
	client, err := NewClient()
	s.NoError(err)

	values := make(url.Values)
	_, err = client.newFormRequest(context.Background(), "/", values)

	s.NoError(err)
}

func (s *Suite) TestClientNewFormRequestWithCrumbs() {
	client, err := NewClient()
	s.NoError(err)

	client.crumbs = &Crumbs{RequestField: "crumbRequestField", Value: "crumb"}

	values := make(url.Values)
	got, err := client.newFormRequest(context.Background(), "/", values)
	s.NoError(err)
	s.Equal(got.Header.Get("crumbRequestField"), "crumb")
}

func (s *Suite) TestClientNewFormRequestError() {
	client, err := NewClient()
	s.NoError(err)

	values := make(url.Values)
	//lint:ignore SA1012 this is a test
	//nolint
	_, err = client.newFormRequest(nil, "/", values)
	s.Error(err)
}

func (s *Suite) TestClientSetCrumbs() {
	s.newMux()
	s.mux.HandleFunc(crumbURL, func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`{"crumb":"crumb"}`,
		))
		s.NoError(err)
	})

	client, err := NewClient(WithBaseURL(s.server.URL), WithUserPassword("admin", "admin"))
	s.NoError(err)

	err = client.setCrumbs(context.Background())
	s.NoError(err)
}

func (s *Suite) TestClientSetCrumbsErrorGet() {
	client, err := NewClient()
	s.NoError(err)

	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)

	s.mux.HandleFunc(crumbURL, func(w http.ResponseWriter, r *http.Request) {})

	//lint:ignore SA1012 this is a test
	//nolint
	err = client.setCrumbs(nil)
	s.Error(err)
}

func (s *Suite) TestClientSetCrumbsErrorUnmarshal() {
	s.newMux()
	s.mux.HandleFunc(crumbURL, func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`{"crumb":"crumb"`,
		))
		s.NoError(err)
	})

	client, err := NewClient(WithBaseURL(s.server.URL), WithUserPassword("admin", "admin"))
	s.NoError(err)

	err = client.setCrumbs(context.Background())
	s.Error(err)
}

func (s *Suite) TestClientPostForm() {
	s.newMux()
	s.addCrumbsHandle()
	s.mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "POST")
		_, err := w.Write([]byte(
			`{"A":"B"}`,
		))
		s.NoError(err)
	})

	client, err := NewClient(WithBaseURL(s.server.URL))
	s.NoError(err)

	type PostBody struct {
		A string `json:"a"`
	}
	_, err = client.postForm(context.Background(), "post", &PostBody{A: "B"})
	s.NoError(err)
}

func (s *Suite) TestClientPostFormCrumbError() {
	s.newMux()

	client, err := NewClient(WithBaseURL(s.server.URL))
	s.NoError(err)

	type PostBody struct {
		A string `json:"a"`
	}

	_, err = client.postForm(context.Background(), "post", &PostBody{A: "B"})
	s.Error(err)
}

func (s *Suite) TestClientPostFormStatusError() {
	s.newMux()
	s.addCrumbsHandle()
	s.mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("500 - Something bad happened!"))
		s.NoError(err)
	})

	client, err := NewClient(WithBaseURL(s.server.URL))
	s.NoError(err)

	type PostBody struct {
		A string `json:"a"`
	}
	_, err = client.postForm(context.Background(), "post", &PostBody{A: "B"})
	s.Error(err)
}

func (s *Suite) TestClientPost() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithUserPassword("admin", "admin"))
	s.NoError(err)

	s.addCrumbsHandle()

	s.mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "POST")
		_, err := w.Write([]byte(
			`<root></root>`,
		))
		s.NoErrorf(err, "w.Write returned %v")
		s.Equal("Basic YWRtaW46YWRtaW4=", r.Header.Get("Authorization"))
	})

	got, err := client.post(context.Background(), "test", nil)
	s.NoError(err)
	s.Equal(http.StatusOK, got.StatusCode)

	all, err := io.ReadAll(got.Body)
	s.NoError(err)
	s.Equal(`<root></root>`, string(all))
}

type brokenXML struct{}

func (b *brokenXML) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	return errors.New("")
}

func (s *Suite) TestClientPostWrongBody() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithUserPassword("admin", "admin"))
	s.NoError(err)

	s.addCrumbsHandle()

	s.mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {})

	_, err = client.post(context.Background(), "test", &brokenXML{})
	s.Error(err)
}

func (s *Suite) TestClientPostNotOK() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithUserPassword("admin", "admin"))
	s.NoError(err)

	s.addCrumbsHandle()

	s.mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "my own error message", http.StatusBadRequest)
	})

	_, err = client.post(context.Background(), "test", nil)
	s.Error(err)
}
