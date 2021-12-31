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
	"strings"
)

// NodeMode represents a Jenkins node mode. Could be either NORMAL or EXCLUSIVE.
type NodeMode string

const (
	// NodeModeNormal sets node usage as "Use this node as much as possible"
	NodeModeNormal NodeMode = "NORMAL"
	// NodeModeExclusive sets node usage as "Only build jobs with label expressions matching this node"
	NodeModeExclusive NodeMode = "EXCLUSIVE"

	// NodesCreateURL is the URL to create a new node
	NodesCreateURL = "/computer/doCreateItem"
	NodesListURL   = "/computer/api/json"
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

type NodesListResponse struct {
	Class          string     `json:"_class"`
	BusyExecutors  int        `json:"busyExecutors"`
	Computer       []Computer `json:"computer"`
	DisplayName    string     `json:"displayName"`
	TotalExecutors int        `json:"totalExecutors"`
}

type AssignedLabels struct {
	Name string `json:"name"`
}

type Executors struct {
}

type LoadStatistics struct {
	Class string `json:"_class"`
}

type SwapSpaceMonitor struct {
	Class                   string `json:"_class"`
	AvailablePhysicalMemory int64  `json:"availablePhysicalMemory"`
	AvailableSwapSpace      int    `json:"availableSwapSpace"`
	TotalPhysicalMemory     int64  `json:"totalPhysicalMemory"`
	TotalSwapSpace          int    `json:"totalSwapSpace"`
}

type TemporarySpaceMonitor struct {
	Class     string `json:"_class"`
	Timestamp int64  `json:"timestamp"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
}

type DiskSpaceMonitor struct {
	Class     string `json:"_class"`
	Timestamp int64  `json:"timestamp"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
}

type ResponseTimeMonitor struct {
	Class     string `json:"_class"`
	Timestamp int64  `json:"timestamp"`
	Average   int    `json:"average"`
}

type ClockMonitor struct {
	Class string `json:"_class"`
	Diff  int    `json:"diff"`
}

type MonitorData struct {
	SwapSpaceMonitor      SwapSpaceMonitor      `json:"hudson.node_monitors.SwapSpaceMonitor"`
	TemporarySpaceMonitor TemporarySpaceMonitor `json:"hudson.node_monitors.TemporarySpaceMonitor"`
	DiskSpaceMonitor      DiskSpaceMonitor      `json:"hudson.node_monitors.DiskSpaceMonitor"`
	ArchitectureMonitor   string                `json:"hudson.node_monitors.ArchitectureMonitor"`
	ResponseTimeMonitor   ResponseTimeMonitor   `json:"hudson.node_monitors.ResponseTimeMonitor"`
	ClockMonitor          ClockMonitor          `json:"hudson.node_monitors.ClockMonitor"`
}

type Computer struct {
	Class               string           `json:"_class"`
	Actions             []interface{}    `json:"actions"`
	AssignedLabels      []AssignedLabels `json:"assignedLabels"`
	Description         string           `json:"description"`
	DisplayName         string           `json:"displayName"`
	Executors           []Executors      `json:"executors"`
	Icon                string           `json:"icon"`
	IconClassName       string           `json:"iconClassName"`
	Idle                bool             `json:"idle"`
	JnlpAgent           bool             `json:"jnlpAgent"`
	LaunchSupported     bool             `json:"launchSupported"`
	LoadStatistics      LoadStatistics   `json:"loadStatistics"`
	ManualLaunchAllowed bool             `json:"manualLaunchAllowed"`
	MonitorData         MonitorData      `json:"monitorData"`
	NumExecutors        int              `json:"numExecutors"`
	Offline             bool             `json:"offline"`
	OfflineCause        interface{}      `json:"offlineCause"`
	OfflineCauseReason  string           `json:"offlineCauseReason"`
	OneOffExecutors     []interface{}    `json:"oneOffExecutors"`
	TemporarilyOffline  bool             `json:"temporarilyOffline"`
	AbsoluteRemotePath  interface{}      `json:"absoluteRemotePath,omitempty"`
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

	resp, err := s.client.PostForm(ctx, NodesCreateURL, nodeRequest)
	if err != nil {
		return nil, resp, err
	}

	return node, resp, nil
}

func (s *NodesService) List(ctx context.Context) ([]Node, *http.Response, error) {
	resp, err := s.client.Get(ctx, NodesListURL)
	if err != nil {
		return nil, resp, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	var listResp NodesListResponse
	err = json.Unmarshal(body, &listResp)
	if err != nil {
		return nil, nil, err
	}

	nodes := make([]Node, len(listResp.Computer))

	for i, node := range listResp.Computer {
		nodes[i] = Node{
			Name:            node.DisplayName,
			NodeDescription: node.Description,
		}
	}

	return nodes, nil, nil
}
