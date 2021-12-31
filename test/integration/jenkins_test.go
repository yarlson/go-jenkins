// Copyright 2021 The go-jenkins AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build integration
// +build integration

package integration

import (
	"context"
	"github.com/yarlson/go-jenkins/jenkins"
	"strings"
	"testing"
)

func TestClientSetCrumbs(t *testing.T) {
	client, err := jenkins.NewClient(jenkins.WithPassword("admin", "admin"))

	if err != nil {
		t.Errorf("getCrumbs() error = %v", err)
		return
	}

	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "TestClientSetCrumbs",
			want:    "Jenkins",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.SetCrumbs(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("getCrumbs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.HasPrefix(client.Crumbs.RequestField, tt.want) {
				t.Errorf("getCrumbs() got = %v, want %v", client.Crumbs.RequestField, tt.want)
			}
		})
	}
}
