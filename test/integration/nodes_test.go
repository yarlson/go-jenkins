// Copyright 2021 The go-jenkins AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/hex"
	"github.com/yarlson/go-jenkins/jenkins"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestNodesCreate(t *testing.T) {
	randBytes := make([]byte, 16)
	rand.Seed(time.Now().UnixNano())
	rand.Read(randBytes)
	name := hex.EncodeToString(randBytes)

	type args struct {
		name string
		node *jenkins.Node
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "create node",
			args: args{
				name: name,
				node: &jenkins.Node{
					Name:               name,
					NodeDescription:    "",
					RemoteFS:           "/var/lib/jenkins",
					NumExecutors:       1,
					Mode:               jenkins.NodeModeExclusive,
					Type:               "hudson.slaves.DumbSlave$DescriptorImpl",
					Labels:             []string{"test"},
					RetentionsStrategy: &jenkins.RetentionsStrategy{StaplerClass: "hudson.slaves.RetentionStrategy$Always"},
				},
			},
			want:    name,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := jenkins.NewClient(jenkins.WithPassword("admin", "admin"))
			if err != nil {
				t.Errorf("Nodes create() error = %v", err)
				return
			}

			got, _, err := client.Nodes.Create(context.Background(), tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("Nodes create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Name, tt.want) {
				t.Errorf("Nodes create() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodesList(t *testing.T) {
	randBytes := make([]byte, 16)
	rand.Seed(time.Now().UnixNano())
	rand.Read(randBytes)
	name := hex.EncodeToString(randBytes)

	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "list nodes",
			want:    name,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := jenkins.NewClient(jenkins.WithPassword("admin", "admin"))
			if err != nil {
				t.Errorf("Nodes list() error = %v", err)
				return
			}

			got, _, err := client.Nodes.List(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Nodes list() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) < 1 {
				t.Errorf("Nodes list() got = %v, want %v", got, tt.want)
			}
		})
	}
}
