package jenkins

import (
	"context"
	"net/http"
)

func (s *Suite) TestNodeFillInNodeDefaults() {
	n := &Node{}
	n.fillInNodeDefaults()

	s.Equal(DefaultJNLPLauncher(), n.Launcher)
	s.Equal(DefaultNodeProperties(), n.NodeProperties)
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
