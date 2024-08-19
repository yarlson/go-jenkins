# go-jenkins

go-jenkins is a Go client library for accessing the Jenkins API. It provides a simple and efficient way to interact with Jenkins programmatically.

⚠️ **Note:** This library is currently in active development and is not yet ready for production use.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [API Documentation](#api-documentation)
- [Contributing](#contributing)
- [Testing](#testing)
- [License](#license)

## Installation

go-jenkins is compatible with modern Go releases in module mode. To install the package, use the following command:

```sh
go get github.com/yarlson/go-jenkins
```

## Usage

To use go-jenkins in your Go code, import it as follows:

```go
import "github.com/yarlson/go-jenkins/jenkins"
```

Here's a basic example of how to use the library to create a new Jenkins node:

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
```

This example demonstrates how to create a new Jenkins client and use it to create a new node.

## API Documentation

For detailed API documentation, please refer to the [GoDoc reference](https://pkg.go.dev/github.com/yarlson/go-jenkins/jenkins).

The library currently supports the following Jenkins API operations:

- Node management (create, list, get, update, delete)
- JNLP and SSH launcher configurations
- Various node properties and configurations

## Contributing

Contributions to go-jenkins are welcome! If you'd like to contribute, please follow these steps:

1. Fork the repository
2. Create a new branch for your feature or bug fix
3. Write your code and tests
4. Ensure all tests pass
5. Submit a pull request

Please make sure to update tests as appropriate and adhere to the existing coding style.

## Testing

To run the tests for go-jenkins, use the following command:

```sh
go test ./...
```

The project uses GitHub Actions for continuous integration. You can view the current test status by clicking on the "Test Status" badge at the top of this README.

## License

go-jenkins is released under the BSD-style license. For more details, see the [LICENSE](LICENSE) file in the repository.

---

For any issues, questions, or suggestions, please [open an issue](https://github.com/yarlson/go-jenkins/issues) on the GitHub repository.
