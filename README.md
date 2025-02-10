# Kommon

Kommon is a Go library for retrieving and managing comments from various sources like GitHub Issues and Pull Requests.

## Features

- Simple and intuitive API
- Support for both GitHub Issues and Pull Requests
- Extensible design for adding new providers

## Installation

```bash
go get github.com/takutakahashi/kommon
```

## Usage

### Command Line Tool

```bash
# Build the CLI tool
go build -o comment ./cmd/comment

# Get comments from a GitHub Issue
export GITHUB_TOKEN=your_github_token
./comment owner repo 123 issue

# Get comments from a GitHub Pull Request
./comment owner repo 456 pr
```

### Library Usage

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/takutakahashi/kommon/pkg/github"
    "github.com/takutakahashi/kommon/pkg/goose"
)

func main() {
    // Create a new GitHub client
    client, err := goose.NewGitHubClient(goose.GitHubOptions{
        Token:  "your-github-token",
        Owner:  "owner",
        Repo:   "repo",
        Number: 123,
        Type:   github.ReferenceTypeIssue,
    })
    if err != nil {
        panic(err)
    }

    // Get comments
    ctx := context.Background()
    comments, err := client.GetComments(ctx)
    if err != nil {
        panic(err)
    }

    // Process comments
    for _, comment := range comments {
        fmt.Printf("Author: %s\n", comment.Author)
        fmt.Printf("Comment: %s\n", comment.Body)
    }
}
```

### Custom Provider Implementation

You can implement your own provider by implementing the `interfaces.CommentProvider` interface:

```go
type CommentProvider interface {
    GetComments(ctx context.Context) ([]Comment, error)
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License