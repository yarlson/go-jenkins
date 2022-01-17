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
	"github.com/stretchr/testify/suite"
	"github.com/yarlson/go-jenkins/jenkins"
	"math/rand"
	"testing"
	"time"
)

type Suite struct {
	suite.Suite
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) TestNodesCreate() {
	randBytes := make([]byte, 16)
	rand.Seed(time.Now().UnixNano())
	rand.Read(randBytes)
	name := hex.EncodeToString(randBytes)

	node := &jenkins.Node{
		Name:               name,
		Description:        "",
		RemoteFS:           "/var/lib/jenkins",
		NumExecutors:       1,
		Mode:               jenkins.NodeModeExclusive,
		Type:               "hudson.slaves.DumbSlave$DescriptorImpl",
		Labels:             []string{"test"},
		RetentionsStrategy: &jenkins.RetentionsStrategy{StaplerClass: "hudson.slaves.RetentionStrategy$Always"},
	}

	client, err := jenkins.NewClient(jenkins.WithPassword("admin", "admin"))
	s.Require().NoError(err)

	got, _, err := client.Nodes.Create(context.Background(), node)
	s.Require().NoError(err)
	s.Equal(name, got.Name)
}

func (s *Suite) TestNodesCreateSSHLauncher() {
	randBytes := make([]byte, 16)
	rand.Seed(time.Now().UnixNano())
	rand.Read(randBytes)
	name := hex.EncodeToString(randBytes)

	node := &jenkins.Node{
		Name:               name,
		Description:        "",
		RemoteFS:           "/var/lib/jenkins",
		NumExecutors:       1,
		Mode:               jenkins.NodeModeExclusive,
		Type:               "hudson.slaves.DumbSlave$DescriptorImpl",
		Labels:             []string{"test"},
		RetentionsStrategy: jenkins.DefaultRetentionsStrategy(),
		Launcher: jenkins.NewSSHLauncher(
			"localhost",
			22,
			"jenkins",
			10,
			15,
			10,
			true,
			jenkins.NewNonVerifyingKeyVerificationStrategy(),
		),
	}

	client, err := jenkins.NewClient(jenkins.WithPassword("admin", "admin"))
	s.Require().NoError(err)

	got, _, err := client.Nodes.Create(context.Background(), node)
	s.Require().NoError(err)
	s.Equal(name, got.Name)
}

func (s *Suite) TestNodesList() {
	client, err := jenkins.NewClient(jenkins.WithPassword("admin", "admin"))
	s.Require().NoError(err)

	got, _, err := client.Nodes.List(context.Background())
	s.Require().NoError(err)
	s.Greater(len(got), 0)
}

func (s *Suite) TestNodesUpdate() {
	randBytes := make([]byte, 16)
	rand.Seed(time.Now().UnixNano())
	rand.Read(randBytes)
	name := hex.EncodeToString(randBytes)

	node := &jenkins.Node{
		Name:               name,
		Description:        "",
		RemoteFS:           "/var/lib/jenkins",
		NumExecutors:       1,
		Mode:               jenkins.NodeModeExclusive,
		Type:               "hudson.slaves.DumbSlave$DescriptorImpl",
		Labels:             []string{"test"},
		RetentionsStrategy: &jenkins.RetentionsStrategy{StaplerClass: "hudson.slaves.RetentionStrategy$Always"},
	}

	client, err := jenkins.NewClient(jenkins.WithPassword("admin", "admin"))
	s.Require().NoError(err)

	got, _, err := client.Nodes.Create(context.Background(), node)
	s.Require().NoError(err)
	s.Equal(name, got.Name)

	node.Description = "updated"
	node, _, err = client.Nodes.Update(context.Background(), node)
	s.Require().NoError(err)
	s.Equal("updated", node.Description)
}
