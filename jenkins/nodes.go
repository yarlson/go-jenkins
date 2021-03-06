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
	// NodesListURL is the URL to list all nodes
	NodesListURL = "/computer/api/json"
	// NodesGetURL is the URL to get a node
	NodesGetURL = "/computer/%s/config.xml"
	// NodesDeleteURL is the URL to delete a node
	NodesDeleteURL = "/computer/%s/doDelete"
)

// Node represents a Jenkins node.
type Node struct {
	XMLName xml.Name `xml:"slave" json:"-"`

	Name               string              `json:"name" xml:"name"`
	Description        string              `json:"nodeDescription" xml:"description"`
	RemoteFS           string              `json:"remoteFS" xml:"remoteFS"`
	NumExecutors       int                 `json:"numExecutors" xml:"numExecutors"`
	Mode               NodeMode            `json:"mode" xml:"mode"`
	Type               NodeType            `json:"type" xml:"type"`
	Labels             Labels              `json:"labelString" xml:"label"`
	RetentionsStrategy *RetentionsStrategy `json:"retentionsStrategy" xml:"retentionsStrategy"`
	Properties         *NodeProperties     `json:"nodeProperties" xml:"nodeProperties"`
	Launcher           Launcher            `json:"launcher" xml:"launcher"`
}

// fillInNodeDefaults fills in default values for the node.
func (n *Node) fillInNodeDefaults() {
	if n.Launcher == nil {
		n.Launcher = DefaultJNLPLauncher()
	}

	if n.Properties == nil {
		n.Properties = DefaultNodeProperties()
	}

	if n.Type == "" {
		n.Type = DefaultNodeType()
	}

	if n.NumExecutors == 0 {
		n.NumExecutors = 1
	}

	if n.RetentionsStrategy == nil {
		n.RetentionsStrategy = DefaultRetentionsStrategy()
	}
}

// UnmarshalXML implements the xml.Unmarshaler interface.
// It decodes the XML attributes into the corresponding struct fields.
// It also decodes the XML child Launcher nodes into the corresponding struct fields.
func (n *Node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias Node // avoids recursive unmarshal
	v := &struct {
		Launcher struct {
			InnerXML []byte `xml:",innerxml"`  // Stores inner XML of the <launcher> element
			Class    string `xml:"class,attr"` // Stores the class name from the <class> attribute
		} `xml:"launcher"`
		*Alias
	}{
		Alias: (*Alias)(n),
	}

	if err := d.DecodeElement(v, &start); err != nil {
		return err
	}

	// Converts InnerXML to a valid XMl document
	launcherXML := []byte(fmt.Sprintf("<root>%s</root>", v.Launcher.InnerXML))

	switch v.Launcher.Class {
	case "hudson.slaves.JNLPLauncher":
		n.Launcher = &JNLPLauncher{}
		err := xml.Unmarshal(launcherXML, n.Launcher)
		if err != nil {
			return err
		}
	case "hudson.slaves.SSHLauncher":
		n.Launcher = &SSHLauncher{}
		err := xml.Unmarshal(launcherXML, n.Launcher)
		if err != nil {
			return err
		}
	}

	return nil
}

// Labels represents Jenkins node labels.
type Labels []string

// MarshalJSON implements the json.Marshaler interface.
// Concatenates all labels with a space.
func (l Labels) MarshalJSON() ([]byte, error) {
	labels := fmt.Sprintf(`"%s"`, strings.Join(l, " "))
	return []byte(labels), nil
}

// RetentionsStrategy represents a Jenkins node retention strategy.
type RetentionsStrategy struct {
	StaplerClass string `json:"stapler-class" xml:"class,attr"`
}

// DefaultRetentionsStrategy represents the default retention strategy.
func DefaultRetentionsStrategy() *RetentionsStrategy {
	return &RetentionsStrategy{StaplerClass: "hudson.slaves.RetentionStrategy$Always"}
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

// NodeType represents a Jenkins node type.
type NodeType string

// DefaultNodeType represents the default Jenkins node type.
func DefaultNodeType() NodeType {
	return "hudson.slaves.DumbSlave$DescriptorImpl"
}

// NodeRequest represents a Jenkins node request.
type NodeRequest struct {
	Name string   `json:"name"`
	Type NodeType `json:"type"`
	JSON string   `json:"json"`
}

// NodesListResponse represents a Jenkins node list response.
type NodesListResponse struct {
	Class          string     `json:"_class"`
	BusyExecutors  int        `json:"busyExecutors"`
	Computer       []Computer `json:"computer"`
	DisplayName    string     `json:"displayName"`
	TotalExecutors int        `json:"totalExecutors"`
}

// Computer represents a Jenkins node/agent.
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

// AssignedLabels represents a Jenkins assigned labels.
type AssignedLabels struct {
	Name string `json:"name"`
}

// Executors represents a Jenkins executors.
type Executors struct {
}

// LoadStatistics represents a Jenkins load statistics.
type LoadStatistics struct {
	Class string `json:"_class"`
}

// MonitorData expose monitoring data
type MonitorData struct {
	SwapSpaceMonitor      SwapSpaceMonitor      `json:"hudson.node_monitors.SwapSpaceMonitor"`
	TemporarySpaceMonitor TemporarySpaceMonitor `json:"hudson.node_monitors.TemporarySpaceMonitor"`
	DiskSpaceMonitor      DiskSpaceMonitor      `json:"hudson.node_monitors.DiskSpaceMonitor"`
	ArchitectureMonitor   string                `json:"hudson.node_monitors.ArchitectureMonitor"`
	ResponseTimeMonitor   ResponseTimeMonitor   `json:"hudson.node_monitors.ResponseTimeMonitor"`
	ClockMonitor          ClockMonitor          `json:"hudson.node_monitors.ClockMonitor"`
}

// SwapSpaceMonitor checks the swap space availability.
type SwapSpaceMonitor struct {
	Class                   string `json:"_class"`
	AvailablePhysicalMemory int64  `json:"availablePhysicalMemory"`
	AvailableSwapSpace      int    `json:"availableSwapSpace"`
	TotalPhysicalMemory     int64  `json:"totalPhysicalMemory"`
	TotalSwapSpace          int    `json:"totalSwapSpace"`
}

// TemporarySpaceMonitor monitors the disk space of "/tmp".
type TemporarySpaceMonitor struct {
	Class     string `json:"_class"`
	Timestamp int64  `json:"timestamp"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
}

// DiskSpaceMonitor checks available disk space of the remote FS root.
type DiskSpaceMonitor struct {
	Class     string `json:"_class"`
	Timestamp int64  `json:"timestamp"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
}

// ResponseTimeMonitor monitors the round-trip response time to this agent.
type ResponseTimeMonitor struct {
	Class     string `json:"_class"`
	Timestamp int64  `json:"timestamp"`
	Average   int    `json:"average"`
}

// ClockMonitor checks clock of a node to detect out of sync clocks.
type ClockMonitor struct {
	Class string `json:"_class"`
	Diff  int    `json:"diff"`
}

// NodesService handles communication with the node related methods of the Jenkins API
type NodesService service

// Create creates a new Jenkins node.
func (s *NodesService) Create(ctx context.Context, node *Node) (*Node, *http.Response, error) {
	node.fillInNodeDefaults()

	str, err := json.Marshal(node)
	if err != nil {
		return nil, nil, err
	}

	nodeRequest := &NodeRequest{
		Name: node.Name,
		Type: node.Type,
		JSON: string(str),
	}

	resp, err := s.client.postForm(ctx, NodesCreateURL, nodeRequest)
	if err != nil {
		return nil, resp, err
	}

	return node, resp, nil
}

// List returns a list of Jenkins nodes.
func (s *NodesService) List(ctx context.Context) ([]Node, *http.Response, error) {
	resp, err := s.client.get(ctx, NodesListURL)
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
			Name:        node.DisplayName,
			Description: node.Description,
		}
	}

	return nodes, nil, nil
}

// Get returns a Jenkins node.
func (s *NodesService) Get(ctx context.Context, name string) (*Node, *http.Response, error) {
	resp, err := s.client.get(ctx, fmt.Sprintf(NodesGetURL, name))
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

	// Golang cannot unmarshal xml 1.1 documents.
	// Jenkins XMLs are basically 1.0 documents.
	body = []byte(strings.Replace(string(body), "xml version=\"1.1\"", "xml version=\"1.0\"", -1))

	var node Node
	err = xml.Unmarshal(body, &node)
	if err != nil {
		return nil, resp, err
	}

	return &node, nil, nil
}

// Update updates a Jenkins node.
func (s *NodesService) Update(ctx context.Context, node *Node) (*Node, *http.Response, error) {
	resp, err := s.client.post(ctx, fmt.Sprintf(NodesGetURL, node.Name), node)
	if err != nil {
		return nil, resp, err
	}

	return node, nil, nil
}

// Delete deletes a Jenkins node.
func (s *NodesService) Delete(ctx context.Context, name string) (*http.Response, error) {
	return s.client.post(ctx, fmt.Sprintf(NodesDeleteURL, name), nil)
}
