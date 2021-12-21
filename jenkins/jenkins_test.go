// Copyright 2021 The go-jenkins AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jenkins

import (
	"context"
	"io"
	"net/url"
	"reflect"
	"testing"
)

func TestConvertBodyStruct(t *testing.T) {
	type args struct {
		body interface{}
	}

	type testBody struct {
		Name string `json:"name"`
	}

	type testBodyNoJson struct {
		Name string
	}

	tests := []struct {
		name string
		args args
		want url.Values
	}{
		{
			name: "TestConvertBodyStruct",
			args: args{body: &testBody{Name: "test"}},
			want: url.Values{"name": []string{"test"}},
		},
		{
			name: "TestConvertBodyStruct no json",
			args: args{body: &testBodyNoJson{Name: "test"}},
			want: url.Values{"Name": []string{"test"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertBodyStruct(tt.args.body)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertBodyStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientNewRequest(t *testing.T) {
	client := NewClient(nil, "http://localhost:8080/", "admin", "admin")

	type args struct {
		ctx    context.Context
		method string
		path   string
		body   io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "TestClientNewRequest",
			args:    args{ctx: context.Background(), method: "GET", path: "/", body: nil},
			want:    "Basic YWRtaW46YWRtaW4=",
			wantErr: false,
		},
		{
			name:    "TestClientNewRequestError",
			args:    args{ctx: nil, method: "GET", path: "/", body: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.NewRequest(tt.args.ctx, tt.args.method, tt.args.path, tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Header.Get("Authorization") != tt.want {
				t.Errorf("NewRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}
