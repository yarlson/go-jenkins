# go-jenkins

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/yarlson/go-jenkins/jenkins)
[![Test Coverage](https://codecov.io/gh/yarlson/go-jenkins/branch/main/graph/badge.svg?token=RFNTCUV32H)](https://codecov.io/gh/yarlson/go-jenkins)
[![Test Status](https://github.com/yarlson/go-jenkins/workflows/tests/badge.svg)](https://github.com/yarlson/go-jenkins/actions?query=workflow%3Atests)

go-jenkins is a Go client library for accessing the Jenkins API.

❗️❗️❗️ The library is in active development and is not yet ready for use!

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
		jenkins.WithUserPassword("admin", "admin"),
	)
	if err != nil {
		panic(err)
	}

	node, _, err := client.Nodes.Create(context.Background(), &jenkins.Node{
		Name:            "test-node",
		Description: "",
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