// Copyright 2021 The go-jenkins AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jenkins

import (
	"context"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
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
	_, err := NewClient(WithPassword("test", "test"))
	s.NoError(err)
}

func (s *Suite) TestNewClientWithToken() {
	_, err := NewClient(WithToken("test", "test"))
	s.NoError(err)
}

func (s *Suite) TestNewClientWithTokenAndPassword() {
	_, err := NewClient(WithToken("test", "test"), WithPassword("test", "test"))
	s.Error(err)
}

func (s *Suite) TestClientNewRequest() {
	client, err := NewClient()
	s.NoError(err)

	_, err = client.NewRequest(context.Background(), "GET", "/", nil)

	s.NoError(err)
}

func (s *Suite) TestClientNewRequestError() {
	client, err := NewClient()
	s.NoError(err)

	//lint:ignore SA1012 this is a test
	_, err = client.NewRequest(nil, "GET", "/", nil)
	s.Error(err)
}

func (s *Suite) testMethod(r *http.Request, want string) {
	s.Equal(want, r.Method)
}

func (s *Suite) TestClientGet() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`{"A":"a"}`,
		))
		s.NoErrorf(err, "w.Write returned %v")
		s.Equal("Basic YWRtaW46YWRtaW4=", r.Header.Get("Authorization"))
	})

	got, err := client.Get(context.Background(), "test")
	s.NoError(err)
	s.Equal(got.StatusCode, http.StatusOK)

	all, err := io.ReadAll(got.Body)
	s.NoError(err)
	s.Equal(string(all), `{"A":"a"}`)
}

func (s *Suite) TestClientGetNotFound() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL))
	s.NoError(err)

	got, err := client.Get(context.Background(), "test_error")
	s.Error(err)
	s.Equal(got.StatusCode, http.StatusNotFound)
}

func (s *Suite) TestClientGetErrorContext() {
	client, err := NewClient()
	s.NoError(err)

	//lint:ignore SA1012 this is a test
	_, err = client.Get(nil, "test_error")
	s.Error(err)
}

func (s *Suite) TestClientGetErrorResponse() {
	client, err := NewClient()
	s.NoError(err)

	deadCtx, cancel := context.WithDeadline(context.Background(), time.Now())
	defer cancel()

	_, err = client.Get(deadCtx, "test_error")
	s.Error(err)
}

func (s *Suite) TestClientGetCookie() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.mux.HandleFunc("/test_cookie", func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		w.Header().Set("Set-Cookie", "test=cookie")
	})

	got, err := client.Get(context.Background(), "test_cookie")
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
	_, err = client.NewFormRequest(context.Background(), "/", values)

	s.NoError(err)
}

func (s *Suite) TestClientNewFormRequestWithCrumbs() {
	client, err := NewClient()
	s.NoError(err)

	client.Crumbs = &Crumbs{RequestField: "crumbRequestField", Value: "crumb"}

	values := make(url.Values)
	got, err := client.NewFormRequest(context.Background(), "/", values)
	s.NoError(err)
	s.Equal(got.Header.Get("crumbRequestField"), "crumb")
}

func (s *Suite) TestClientNewFormRequestError() {
	client, err := NewClient()
	s.NoError(err)

	values := make(url.Values)
	//lint:ignore SA1012 this is a test
	_, err = client.NewFormRequest(nil, "/", values)
	s.Error(err)
}

func (s *Suite) TestClientSetCrumbs() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.mux.HandleFunc(crumbURL, func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`{"crumb":"crumb"}`,
		))
		s.NoError(err)
	})

	got, err := client.SetCrumbs(context.Background())
	s.NoError(err)
	s.Equal(got.StatusCode, http.StatusOK)
}

func (s *Suite) TestClientSetCrumbsErrorGet() {
	client, err := NewClient()
	s.NoError(err)

	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)

	s.mux.HandleFunc(crumbURL, func(w http.ResponseWriter, r *http.Request) {})

	//lint:ignore SA1012 this is a test
	_, err = client.SetCrumbs(nil)
	s.Error(err)
}

func (s *Suite) TestClientSetCrumbsErrorUnmarshal() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.mux.HandleFunc(crumbURL, func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`{"crumb":"crumb"`,
		))
		s.NoError(err)
	})

	_, err = client.SetCrumbs(context.Background())
	s.Error(err)
}

func (s *Suite) TestClientPostForm() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL))
	s.NoError(err)

	type PostBody struct {
		A string `json:"a"`
	}

	s.mux.HandleFunc(crumbURL, func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`{"crumbRequestField":"crumb", "crumb":"crumb"}`,
		))
		s.NoError(err)
	})

	s.mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "POST")
		_, err := w.Write([]byte(
			`{"A":"B"}`,
		))
		s.NoError(err)
	})

	_, err = client.PostForm(context.Background(), "post", &PostBody{A: "B"})
	s.NoError(err)
}

func (s *Suite) TestClientPostFormCrumbError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL))
	s.NoError(err)

	type PostBody struct {
		A string `json:"a"`
	}

	_, err = client.PostForm(context.Background(), "post", &PostBody{A: "B"})
	s.Error(err)
}

func (s *Suite) TestClientPostFormStatusError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL))
	s.NoError(err)

	type PostBody struct {
		A string `json:"a"`
	}

	s.mux.HandleFunc(crumbURL, func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`{"crumbRequestField":"crumb", "crumb":"crumb"}`,
		))
		s.NoError(err)
	})

	s.mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("500 - Something bad happened!"))
		s.NoError(err)
	})
	_, err = client.PostForm(context.Background(), "post", &PostBody{A: "B"})
	s.Error(err)
}
