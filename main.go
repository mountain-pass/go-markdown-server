package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type Server struct {
	contentDir           string
	port                string
	enableSecurityHeaders bool
}

func NewServer(contentDir, port string, enableSecurityHeaders bool) *Server {
	return &Server{
		contentDir:           contentDir,
		port:                port,
		enableSecurityHeaders: enableSecurityHeaders,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.securityHeadersMiddleware(s.handleMarkdown))
	
	fmt.Printf("Starting server on port %s, serving content from %s\n", s.port, s.contentDir)
	return http.ListenAndServe(":"+s.port, nil)
}

// securityHeadersMiddleware adds security headers to all responses if enabled
func (s *Server) securityHeadersMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only add security headers if enabled
		if s.enableSecurityHeaders {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
			
			// Content Security Policy - allowing iframe embedding as requested
			// Note: Omitting X-Frame-Options since user wants iframe support
			csp := "default-src 'self'; " +
				"style-src 'self' 'unsafe-inline'; " +
				"script-src 'self'; " +
				"img-src 'self' data: https:; " +
				"font-src 'self'; " +
				"connect-src 'self'; " +
				"frame-ancestors *; " + // Allow iframe embedding
				"base-uri 'self'"
			w.Header().Set("Content-Security-Policy", csp)
		}
		
		// Call the next handler
		next(w, r)
	}
}

func (s *Server) handleMarkdown(w http.ResponseWriter, r *http.Request) {
	// Clean the URL path
	urlPath := strings.TrimPrefix(r.URL.Path, "/")
	if urlPath == "" {
		urlPath = "index.md"
	}
	
	// Security: Validate and sanitize the path to prevent directory traversal
	if err := s.validatePath(urlPath); err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	
	// Handle CSS file requests
	if urlPath == "style.css" {
		cssPath := filepath.Join(s.contentDir, "style.css")
		// Security: Ensure the resolved path is still within content directory
		if !s.isPathSafe(cssPath) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		if _, err := os.Stat(cssPath); err == nil {
			w.Header().Set("Content-Type", "text/css")
			http.ServeFile(w, r, cssPath)
			return
		}
		http.NotFound(w, r)
		return
	}
	
	// Add .md extension if not present and not a directory
	if !strings.HasSuffix(urlPath, ".md") && !strings.HasSuffix(urlPath, "/") {
		urlPath += ".md"
	}
	
	filePath := filepath.Join(s.contentDir, urlPath)
	
	// Security: Ensure the resolved path is still within content directory
	if !s.isPathSafe(filePath) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Try with index.md if it's a directory
		if strings.HasSuffix(urlPath, "/") {
			indexPath := filepath.Join(s.contentDir, urlPath, "index.md")
			if !s.isPathSafe(indexPath) {
				http.Error(w, "Invalid path", http.StatusBadRequest)
				return
			}
			filePath = indexPath
		} else {
			// If the requested file doesn't exist, try to serve index.md instead
			indexPath := filepath.Join(s.contentDir, "index.md")
			if !s.isPathSafe(indexPath) {
				http.Error(w, "Invalid path", http.StatusBadRequest)
				return
			}
			if _, indexErr := os.Stat(indexPath); indexErr == nil {
				filePath = indexPath
			} else {
				http.NotFound(w, r)
				return
			}
		}
	}
	
	// Read markdown file
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	
	// Convert markdown to HTML
	htmlContent := s.markdownToHTML(content)
	
	// Render with template
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/style.css">
</head>
<body>
    <div class="container">
        <nav>
            <a href="/">Home</a>
        </nav>
        <main>
            {{.Content}}
        </main>
    </div>
</body>
</html>`
	
	t, err := template.New("page").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	
	data := struct {
		Title   string
		Content template.HTML
	}{
		Title:   s.extractTitle(string(content)),
		Content: template.HTML(htmlContent),
	}
	
	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) markdownToHTML(md []byte) string {
	// Create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	
	// Create HTML renderer with options
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	
	// Parse and render
	doc := p.Parse(md)
	return string(markdown.Render(doc, renderer))
}

func (s *Server) extractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return "Markdown Server"
}

func (s *Server) ensureSampleContent() error {
	// Check if content directory is empty
	isEmpty, err := s.isContentDirEmpty()
	if err != nil {
		return err
	}
	
	if isEmpty {
		// Create sample index.md file
		indexPath := filepath.Join(s.contentDir, "index.md")
		sampleContent := `# Welcome to the Markdown Server

This is a sample markdown file that demonstrates the functionality of our Go-based markdown server.

## Features

- **Markdown to HTML conversion**: All ` + "`" + `.md` + "`" + ` files are automatically converted to HTML
- **Clean URLs**: Access files with or without the ` + "`" + `.md` + "`" + ` extension
- **Template rendering**: Content is wrapped in a clean HTML template
- **CSS styling**: Styles are served from the content directory
- **Auto-generated content**: This sample file was created automatically!

## Getting Started

1. Place your markdown files in the ` + "`" + `content/` + "`" + ` directory
2. Start the server
3. Navigate to ` + "`" + `http://localhost:8080` + "`" + ` to view your content

## Sample Content

Here's some sample markdown content:

### Code Example

` + "```" + `go
func main() {
    fmt.Println("Hello, Markdown Server!")
}
` + "```" + `

### Lists

- Item 1
- Item 2
- Item 3

### Links

Visit [GitHub](https://github.com) for more projects.

---

*This server automatically converts this markdown to HTML!*
`

		if err := os.WriteFile(indexPath, []byte(sampleContent), 0644); err != nil {
			return fmt.Errorf("failed to create sample index.md: %w", err)
		}
		
		// Create style.css file if it doesn't exist
		cssPath := filepath.Join(s.contentDir, "style.css")
		if _, err := os.Stat(cssPath); os.IsNotExist(err) {
			cssContent := `/* Reset and base styles */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    line-height: 1.6;
    color: #333;
    background-color: #f8f9fa;
}

.container {
    max-width: 800px;
    margin: 0 auto;
    background-color: white;
    min-height: 100vh;
    box-shadow: 0 0 20px rgba(0, 0, 0, 0.1);
}

/* Navigation */
nav {
    background-color: #2c3e50;
    padding: 1rem 2rem;
    border-bottom: 3px solid #3498db;
}

nav a {
    color: white;
    text-decoration: none;
    font-weight: 500;
    font-size: 1.1rem;
}

nav a:hover {
    color: #3498db;
}

/* Main content */
main {
    padding: 2rem;
}

/* Typography */
h1, h2, h3, h4, h5, h6 {
    margin-bottom: 1rem;
    color: #2c3e50;
    line-height: 1.2;
}

h1 {
    font-size: 2.5rem;
    border-bottom: 3px solid #3498db;
    padding-bottom: 0.5rem;
    margin-bottom: 1.5rem;
}

h2 {
    font-size: 2rem;
    margin-top: 2rem;
    color: #34495e;
}

h3 {
    font-size: 1.5rem;
    margin-top: 1.5rem;
    color: #34495e;
}

p {
    margin-bottom: 1rem;
    text-align: justify;
}

/* Links */
a {
    color: #3498db;
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
    color: #2980b9;
}

/* Lists */
ul, ol {
    margin-bottom: 1rem;
    padding-left: 2rem;
}

li {
    margin-bottom: 0.5rem;
}

/* Code blocks */
pre {
    background-color: #f4f4f4;
    border: 1px solid #ddd;
    border-radius: 4px;
    padding: 1rem;
    margin-bottom: 1rem;
    overflow-x: auto;
    font-family: 'Monaco', 'Courier New', monospace;
    font-size: 0.9rem;
}

code {
    background-color: #f4f4f4;
    padding: 0.2rem 0.4rem;
    border-radius: 3px;
    font-family: 'Monaco', 'Courier New', monospace;
    font-size: 0.9rem;
}

pre code {
    background-color: transparent;
    padding: 0;
}

/* Blockquotes */
blockquote {
    border-left: 4px solid #3498db;
    margin: 1rem 0;
    padding: 0.5rem 1rem;
    background-color: #f8f9fa;
    font-style: italic;
}

/* Horizontal rules */
hr {
    border: none;
    border-top: 2px solid #ecf0f1;
    margin: 2rem 0;
}

/* Tables */
table {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: 1rem;
}

th, td {
    border: 1px solid #ddd;
    padding: 0.75rem;
    text-align: left;
}

th {
    background-color: #f8f9fa;
    font-weight: 600;
}

tr:nth-child(even) {
    background-color: #f8f9fa;
}

/* Strong and emphasis */
strong {
    font-weight: 600;
    color: #2c3e50;
}

em {
    font-style: italic;
    color: #34495e;
}

/* Responsive design */
@media (max-width: 768px) {
    .container {
        margin: 0;
        box-shadow: none;
    }
    
    nav {
        padding: 1rem;
    }
    
    main {
        padding: 1rem;
    }
    
    h1 {
        font-size: 2rem;
    }
    
    h2 {
        font-size: 1.5rem;
    }
    
    pre {
        font-size: 0.8rem;
    }
}`

			if err := os.WriteFile(cssPath, []byte(cssContent), 0644); err != nil {
				return fmt.Errorf("failed to create sample style.css: %w", err)
			}
			
			fmt.Printf("Created sample style.css file at %s\n", cssPath)
		}
		
		fmt.Printf("Created sample index.md file at %s\n", indexPath)
	}
	
	return nil
}

func (s *Server) isContentDirEmpty() (bool, error) {
	entries, err := os.ReadDir(s.contentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	
	// Check if there are any .md files
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			return false, nil
		}
	}
	
	return true, nil
}

// validatePath checks for obvious path traversal attempts
func (s *Server) validatePath(path string) error {
	// Check for path traversal patterns
	if strings.Contains(path, "..") ||
		strings.Contains(path, "//") ||
		strings.HasPrefix(path, "/") ||
		strings.Contains(path, "\\") {
		return fmt.Errorf("invalid path: contains dangerous characters")
	}
	
	// Only allow alphanumeric, dash, underscore, dot, and slash
	for _, char := range path {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.' || char == '/') {
			return fmt.Errorf("invalid path: contains invalid characters")
		}
	}
	
	return nil
}

// isPathSafe ensures the resolved path is within the content directory
func (s *Server) isPathSafe(requestedPath string) bool {
	// Get absolute paths
	contentAbs, err := filepath.Abs(s.contentDir)
	if err != nil {
		return false
	}
	
	requestedAbs, err := filepath.Abs(requestedPath)
	if err != nil {
		return false
	}
	
	// Check if the requested path is within the content directory
	rel, err := filepath.Rel(contentAbs, requestedAbs)
	if err != nil {
		return false
	}
	
	// If the relative path starts with "..", it's outside the content directory
	return !strings.HasPrefix(rel, "..")
}

func main() {
	contentDir := os.Getenv("CONTENT_DIR")
	if contentDir == "" {
		contentDir = "./content"
	}
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	// Check if security headers should be enabled (default: enabled)
	enableSecurityHeaders := true
	if securityHeadersEnv := os.Getenv("HTTP_SECURITY_HEADERS"); securityHeadersEnv == "disable" {
		enableSecurityHeaders = false
		fmt.Println("HTTP security headers disabled")
	}
	
	// Create content directory if it doesn't exist
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		log.Fatal("Failed to create content directory:", err)
	}
	
	server := NewServer(contentDir, port, enableSecurityHeaders)
	
	// Ensure sample content exists if directory is empty
	if err := server.ensureSampleContent(); err != nil {
		log.Printf("Warning: Failed to create sample content: %v", err)
	}
	
	log.Fatal(server.Start())
}