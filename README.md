# Go Markdown Server

A low memory, <10MB Go web server that serves markdown files converted to HTML.

## Features

- 🔄 **Automatic Markdown to HTML conversion** using the `gomarkdown` library
- 🎨 **Clean, responsive HTML template** with modern CSS styling
- 📁 **File-based routing** - serve `.md` files from the `content/` directory
- 🔗 **Clean URLs** - access files with or without the `.md` extension
- 📦 **Static file serving** for CSS, images, and other assets
- 🐳 **Ultra-minimal Docker containers** using scratch base image (no OS!)
- 🛡️ **Auto-generated sample content** when content directory is empty

## Project Structure

```
go-markdown-server/
├── main.go              # Main server application
├── go.mod              # Go module dependencies
├── go.sum              # Dependency checksums (generated)
├── Dockerfile          # Docker build configuration (scratch-based)
├── LICENSE             # MIT License
├── README.md           # This file
├── content/            # Directory for markdown files
│   └── index.md        # Sample homepage content
└── static/             # Directory for static assets
    └── style.css       # CSS styling for HTML output
```

## Prerequisites

- Go 1.21 or later
- Docker (optional, for containerized deployment)

## Installation & Setup

1. **Install Go dependencies:**
   ```bash
   go mod tidy
   ```

2. **Run the server locally:**
   ```bash
   go run main.go
   ```

3. **Access the server:**
   Open your browser to `http://localhost:8080`

## Configuration

The server can be configured using environment variables:

- `PORT`: Server port (default: `8080`)
- `CONTENT_DIR`: Directory containing markdown files (default: `./content`)
- `HTTP_SECURITY_HEADERS`: Enable/disable HTTP security headers (default: `enable`, set to `disable` to turn off)

Example:
```bash
PORT=3000 CONTENT_DIR=/path/to/markdown/files HTTP_SECURITY_HEADERS=disable go run main.go
```

## Security Features

### HTTP Security Headers

The server includes comprehensive HTTP security headers by default to protect against common web vulnerabilities:

- **X-Content-Type-Options**: Prevents MIME type sniffing attacks
- **X-XSS-Protection**: Enables XSS filtering in browsers  
- **Referrer-Policy**: Controls referrer information sharing
- **X-Permitted-Cross-Domain-Policies**: Blocks Flash/PDF cross-domain requests
- **Content-Security-Policy**: Comprehensive CSP that allows iframe embedding while maintaining security

**Note**: Security headers can be disabled by setting `HTTP_SECURITY_HEADERS=disable` if needed for compatibility with legacy systems.

### Container Security

The Docker deployment includes advanced security hardening:

- **Minimal privileges**: All Linux capabilities dropped except what's necessary
- **Non-root execution**: Runs as user `1001:1001` inside the container
- **No privilege escalation**: `no-new-privileges` security option enabled
- **Path traversal protection**: Server validates all file paths to prevent directory traversal attacks

### Input Validation

- **Path sanitization**: All URL paths are validated and sanitized
- **File extension restrictions**: Only serves `.md` and `.css` files from the content directory
- **Directory containment**: Server ensures all file access stays within the designated content directory

## Docker Deployment

### Ultra-Minimal Container (Recommended)

This Dockerfile uses a `scratch` base image, creating an extremely small container with just the Go binary and essential files - no operating system!

**Benefits:**
- **Tiny size**: ~10-15MB total (vs ~50MB+ with Alpine)
- **Enhanced security**: No shell, package manager, or OS vulnerabilities
- **Fast startup**: Minimal overhead
- **Production-ready**: Statically linked binary with all dependencies included

### Using Docker Compose (Recommended)

**Build and run with Docker Compose:**
```bash
docker-compose up --build
```

**Build and run in background:**
```bash
docker-compose up --build -d
```

**Run development version:**
```bash
docker-compose --profile dev up --build
```

**Build the image without running:**
```bash
docker-compose build
```

**Stop services:**
```bash
docker-compose down
```

### Publishing the Image

**Build and tag for publishing:**
```bash
docker-compose build
docker tag mountainpass/go-markdown-server:latest mountainpass/go-markdown-server:v1.0.0
```

**Push to registry:**
```bash
docker push mountainpass/go-markdown-server:latest
docker push mountainpass/go-markdown-server:v1.0.0
```

**Pull and run from registry:**
```bash
docker pull mountainpass/go-markdown-server:latest
docker run -p 8080:8080 mountainpass/go-markdown-server:latest
```

### Manual Docker Build (Alternative)

### Build the Docker image:
```bash
docker build -t mountainpass/go-markdown-server .
```

### Run the container:
```bash
docker run -p 8080:8080 mountainpass/go-markdown-server
```

### Run with custom content directory:
```bash
docker run -p 8080:8080 -v /path/to/your/content:/content mountainpass/go-markdown-server
```

### Environment variables in Docker:
```bash
docker run -p 3000:3000 -e PORT=3000 -e CONTENT_DIR=/content mountainpass/go-markdown-server
```

## Usage

1. **Add markdown files** to the `content/` directory
2. **Access files** via clean URLs:
   - `http://localhost:8080/` → serves `content/index.md`
   - `http://localhost:8080/about` → serves `content/about.md`
   - `http://localhost:8080/docs/setup` → serves `content/docs/setup.md`

3. **Static assets** can be placed in the `static/` directory and accessed via `/static/` URL path

4. **Auto-generated content**: If the content directory is empty, a sample `index.md` is automatically created

## Markdown Features Supported

- Headers (H1-H6)
- Lists (ordered and unordered)
- Code blocks with syntax highlighting
- Links and images
- Tables
- Blockquotes
- Horizontal rules
- **Bold** and *italic* text
- Automatic heading IDs for anchor links

## Development

To modify the server:

1. Edit `main.go` for server logic changes
2. Edit `static/style.css` for styling changes
3. Add sample content to `content/` directory

The server automatically extracts page titles from the first H1 header (`# Title`) in each markdown file.

## Container Security & Performance

The scratch-based Docker image provides:
- **No attack surface**: No shell, package manager, or OS components
- **Minimal size**: Only contains the statically-linked Go binary and essential files
- **Fast deployment**: Smaller images mean faster pulls and starts
- **Production security**: No unnecessary system packages or vulnerabilities

**Note**: The scratch approach works because Go can compile to a statically-linked binary that includes all dependencies. This is ideal for microservices and containerized deployments.

## License

This project is open source and available under the MIT License.