package jenkins

import (
	"context"
	"fmt"
	"net/http"
)

func (s *Suite) TestNodeFillInNodeDefaults() {
	n := &Node{}
	n.fillInNodeDefaults()

	s.Equal(DefaultJNLPLauncher(), n.Launcher)
	s.Equal(DefaultNodeProperties(), n.Properties)
	s.Equal(DefaultNodeType(), n.Type)
	s.Equal(1, n.NumExecutors)
	s.Equal(DefaultRetentionsStrategy(), n.RetentionsStrategy)
}

func (s *Suite) TestNodesServiceCreate() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.addCrumbsHandle()

	s.mux.HandleFunc(NodesCreateURL, func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "POST")
		_, err := w.Write([]byte(
			`{"A":"B"}`,
		))
		s.NoError(err)
	})

	_, _, err = client.Nodes.Create(context.Background(), &Node{})
	s.NoError(err)
}

func (s *Suite) TestNodesServiceCreateError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.addCrumbsHandle()

	_, _, err = client.Nodes.Create(context.Background(), &Node{})
	s.Error(err)
}

func (s *Suite) TestNodesServiceList() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.addCrumbsHandle()

	s.mux.HandleFunc(NodesListURL, func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`{"computer":[{"displayName": "test"}]}`,
		))
		s.NoError(err)
	})
	nodes, _, err := client.Nodes.List(context.Background())
	s.NoError(err)
	s.Equal([]Node{{Name: "test"}}, nodes)
}

func (s *Suite) TestNodesServiceListError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	//lint:ignore SA1012 this is a test
	_, _, err = client.Nodes.List(nil)
	s.Error(err)
}

func (s *Suite) TestNodesServiceListUnmarshalError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.addCrumbsHandle()

	s.mux.HandleFunc(NodesListURL, func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`{"computer":[{"displayName": `,
		))
		s.NoError(err)
	})
	_, _, err = client.Nodes.List(context.Background())
	s.Error(err)
}

func (s *Suite) TestNodesServiceGet() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.mux.HandleFunc(fmt.Sprintf(NodesGetURL, "test"), func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`
<?xml version="1.1" encoding="UTF-8"?>
<slave>
  <name>test</name>
  <description></description>
  <remoteFS>/var/lib/jenkins</remoteFS>
  <numExecutors>1</numExecutors>
  <mode>EXCLUSIVE</mode>
  <retentionStrategy class="hudson.slaves.RetentionStrategy$Always"/>
  <launcher class="hudson.slaves.JNLPLauncher">
    <workDirSettings>
      <disabled>false</disabled>
      <internalDir>remoting</internalDir>
      <failIfWorkDirIsMissing>false</failIfWorkDirIsMissing>
    </workDirSettings>
    <webSocket>false</webSocket>
  </launcher>
  <label>test</label>
  <nodeProperties/>
</slave>
`,
		))
		s.NoError(err)
	})

	node, _, err := client.Nodes.Get(context.Background(), "test")

	s.NoError(err)
	s.Equal("test", node.Name)
	s.IsType(&JNLPLauncher{}, node.Launcher)
}

func (s *Suite) TestNodesServiceGetError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	_, _, err = client.Nodes.Get(context.Background(), "test")

	s.Error(err)
}

func (s *Suite) TestNodesServiceGetUnmarshalError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.mux.HandleFunc(fmt.Sprintf(NodesGetURL, "test"), func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`
<?xml version="2.0" encoding="UTF-8"?>
<slave>

</slave>
`,
		))
		s.NoError(err)
	})

	_, _, err = client.Nodes.Get(context.Background(), "test")

	s.Error(err)
}

func (s *Suite) TestNodesServiceUpdate() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.addCrumbsHandle()

	s.mux.HandleFunc(fmt.Sprintf(NodesGetURL, "test"), func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "POST")
	})

	_, _, err = client.Nodes.Update(context.Background(), &Node{
		Name:        "test",
		Description: "",
		RemoteFS:    "/var/lib/jenkins",
		Mode:        NodeModeExclusive,
		Labels:      []string{"test"},
	})

	s.NoError(err)
}

func (s *Suite) TestNodesServiceUpdateError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	_, _, err = client.Nodes.Update(context.Background(), &Node{
		Name:        "test",
		Description: "",
		RemoteFS:    "/var/lib/jenkins",
		Mode:        NodeModeExclusive,
		Labels:      []string{"test"},
	})

	s.Error(err)
}
