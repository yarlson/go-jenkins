# go-jenkins

go-jenkins is a Go client library for accessing the Jenkins API.

## Installation
go-jenkins is compatible with modern Go releases in module mode, with Go installed:
```shell
go get github.com/yarlson/go-jenkins
```

## Usage
```go
import "github.com/yarlson/go-jenkins/jenkins"	
```

## Example
```go
package main

import (
	"context"
	"fmt"
	"github.com/yarlson/go-jenkins/jenkins"
)

func main() {
	client, err := jenkins.NewClient(
		jenkins.WithBaseURL("http://localhost:8080"),
		jenkins.WithPassword("admin", "admin"),
	)
	if err != nil {
		panic(err)
	}

	node, _, err := client.Nodes.Create(context.Background(), &jenkins.Node{
		Name:            "test-node",
		NodeDescription: "",
		RemoteFS:        "/var/lib/jenkins",
		NumExecutors:    1,
		Mode:            jenkins.NodeModeExclusive,
		Labels:          []string{"test"},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(node)
}

```