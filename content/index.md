# Welcome to the Markdown Server

This is a sample markdown file that demonstrates the functionality of our Go-based markdown server.

## Features

- **Markdown to HTML conversion**: All `.md` files are automatically converted to HTML
- **Clean URLs**: Access files with or without the `.md` extension
- **Template rendering**: Content is wrapped in a clean HTML template
- **Static file serving**: CSS and other assets are served from the `/static/` path

## Getting Started

1. Place your markdown files in the `content/` directory
2. Start the server
3. Navigate to `http://localhost:8080` to view your content

## Sample Content

Here's some sample markdown content:

### Code Example

```go
func main() {
    fmt.Println("Hello, Markdown Server!")
}
```

### Lists

- Item 1
- Item 2
- Item 3

### Links

Visit [GitHub](https://github.com) for more projects.

---

*This server automatically converts this markdown to HTML!*