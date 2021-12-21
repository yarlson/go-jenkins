// Copyright 2021 The go-jenkins AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jenkins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// NodeMode represents a Jenkins node mode. Could be either NORMAL or EXCLUSIVE.
type NodeMode string

const (
	// NodeModeNormal sets node usage as "Use this node as much as possible"
	NodeModeNormal NodeMode = "NORMAL"
	// NodeModeExclusive sets node usage as "Only build jobs with label expressions matching this node"
	NodeModeExclusive NodeMode = "EXCLUSIVE"

	// NodeCreateURL is the URL to create a new node
	NodeCreateURL = "/computer/doCreateItem"
)

// Labels represents Jenkins node labels.
type Labels []string

// MarshalJSON implements the json.Marshaler interface.
// Concatenates all labels with a space.
func (l Labels) MarshalJSON() ([]byte, error) {
	labels := fmt.Sprintf(`"%s"`, strings.Join(l, " "))
	return []byte(labels), nil
}

// Node represents a Jenkins node.
type Node struct {
	Name               string              `json:"name"`
	NodeDescription    string              `json:"nodeDescription"`
	RemoteFS           string              `json:"remoteFS"`
	NumExecutors       int                 `json:"numExecutors"`
	Mode               NodeMode            `json:"mode"`
	Type               string              `json:"type"`
	Labels             Labels              `json:"labelString"`
	RetentionsStrategy *RetentionsStrategy `json:"retentionsStrategy"`
	NodeProperties     *NodeProperties     `json:"nodeProperties"`
	Launcher           interface{}         `json:"launcher"`
}

// RetentionsStrategy represents a Jenkins node retention strategy.
type RetentionsStrategy struct {
	StaplerClass string `json:"stapler-class"`
}

// NodeProperties represents a Jenkins node properties.
type NodeProperties struct {
	StaplerClassBag string `json:"stapler-class-bag"`
}

// DefaultNodeProperties returns the default node properties.
func DefaultNodeProperties() *NodeProperties {
	return &NodeProperties{
		StaplerClassBag: "true",
	}
}

// JNLPLauncher represents a Jenkins JNLP launcher.
type JNLPLauncher struct {
	StaplerClass string `json:"stapler-class"`
}

// DefaultJNLPLauncher returns the default JNLP launcher.
func DefaultJNLPLauncher() *JNLPLauncher {
	return &JNLPLauncher{
		StaplerClass: "hudson.slaves.JNLPLauncher",
	}
}

// NodeRequest represents a Jenkins node request.
type NodeRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
	JSON string `json:"json"`
}

type NodesService service

// Create creates a new Jenkins node.
func (s *NodesService) Create(ctx context.Context, node *Node) (*Node, *http.Response, error) {
	if node.Launcher == nil {
		node.Launcher = DefaultJNLPLauncher()
	}

	if node.NodeProperties == nil {
		node.NodeProperties = DefaultNodeProperties()
	}

	str, err := json.Marshal(node)
	if err != nil {
		return nil, nil, err
	}

	nodeRequest := &NodeRequest{
		Name: node.Name,
		Type: node.Type,
		JSON: string(str),
	}

	if _, err := s.client.SetCrumbs(ctx); err != nil {
		return nil, nil, err
	}

	resp, err := s.client.PostForm(ctx, NodeCreateURL, nodeRequest)
	if err != nil {
		return nil, resp, err
	}

	return node, resp, nil
}
