package main

import (
	"context"
	"fmt"
	"github.com/yarlson/go-jenkins/jenkins"
)

func main() {
	client, err := jenkins.NewClient(
		jenkins.WithBaseURL("http://localhost:8080"),
		jenkins.WithUserPassword("admin", "admin"),
	)
	if err != nil {
		panic(err)
	}

	node, _, err := client.Nodes.Create(context.Background(), &jenkins.Node{
		Name:         "test-node",
		Description:  "",
		RemoteFS:     "/var/lib/jenkins",
		NumExecutors: 1,
		Mode:         jenkins.NodeModeExclusive,
		Labels:       []string{"test"},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(node)
}
