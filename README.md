# Multitool Server

A lightweight HTTP server written in Go that provides various utility endpoints for testing, debugging, and learning purposes. This server is designed to be a Swiss Army knife for HTTP-based testing scenarios.

## 🚀 Features

- **Network Information**: Get hostname, IP addresses, and source IP
- **Environment Inspection**: View environment variables and headers
- **File System Access**: Browse directory contents with path traversal protection
- **HTTP Testing**: Make external HTTP requests and simulate errors
- **Robust Error Handling**: All endpoints return clear error messages and log failures
- **Input Validation**: `/req` endpoint validates URLs and only allows http/https schemes
- **Docker Ready**: Pre-built Docker image available
- **Kubernetes Compatible**: Ready for container orchestration

## 📋 API Endpoints

### Information Endpoints

| Endpoint | Description | Example |
|----------|-------------|---------|
| `/` | Returns hostname, IP address, and optional custom text | `GET /` |
| `/hello` | Same as root endpoint | `GET /hello` |
| `/host` | Returns the hostname | `GET /host` |
| `/ip` | Returns the IP address of the host | `GET /ip` |
| `/source` | Returns the source IP address of the request | `GET /source` |

### Environment & Headers

| Endpoint | Description | Example |
|----------|-------------|---------|
| `/env` | Returns all environment variables | `GET /env` |
| `/headers` | Returns all request headers | `GET /headers` |

### File System

| Endpoint | Description | Example |
|----------|-------------|---------|
| `/ls` | Lists directory contents | `GET /ls?path=/tmp` |
| `/ls` | Lists current directory (default) | `GET /ls` |

### HTTP Testing

| Endpoint | Description | Example |
|----------|-------------|---------|
| `/req` | Makes HTTP request to external URL (http/https only, with validation) | `GET /req?url=https://httpbin.org/get` |
| `/req` | Defaults to Google if no URL provided | `GET /req` |
| `/error` | Always returns HTTP 500 error | `GET /error` |
| `/error2` | Returns HTTP 500 every second request | `GET /error2` |

## 🛡️ Security Features

- **Path Traversal Protection**: The `/ls` endpoint is protected against directory traversal attacks
- **Input Validation**: All user inputs are validated and sanitized
- **/req URL Validation**: Only valid `http` and `https` URLs are allowed; invalid or unsupported schemes return a `400 Bad Request` or `502 Bad Gateway` with a clear error message
- **Error Handling**: All endpoints return clear error messages and log failures

## ⚠️ Error Responses & Logging

- All endpoints log errors to the server log for debugging and auditing.
- If a response cannot be written (e.g., client disconnects), the error is logged and the handler returns early.
- The `/req` endpoint:
  - Returns `400 Bad Request` for invalid or unsupported URLs (e.g., missing scheme, not http/https)
  - Returns `502 Bad Gateway` for network or remote server errors
  - Logs all errors with context (invalid input, failed requests, etc.)

**Example error response for invalid URL:**
```bash
$ curl -i 'http://localhost:8080/req?url=not_a_url'
HTTP/1.1 502 Bad Gateway
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: ...
Content-Length: ...

Failed to make request: Get "not_a_url": unsupported protocol scheme ""
```

**Example error response for unsupported scheme:**
```bash
$ curl -i 'http://localhost:8080/req?url=ftp://example.com/file.txt'
HTTP/1.1 502 Bad Gateway
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: ...
Content-Length: ...

Failed to make request: Get "ftp://example.com/file.txt": unsupported protocol scheme "ftp"
```

## 🐳 Docker

### Quick Start

```bash
# Pull the image
docker pull przemekmalak/multitoolserver

# Run the container
docker run -d -p 8080:8080 --name multitoolserver przemekmalak/multitoolserver

# Test the server
curl http://localhost:8080/hello
```

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd multitoolserver

# Build the image (AMD64 architecture)
docker build --platform linux/amd64 -t multitoolserver .

# Or using buildx for multi-platform support
docker buildx build --platform linux/amd64,linux/arm64 -t multitoolserver --load .

# Run the container
docker run -d -p 8080:8080 --name multitoolserver multitoolserver
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `RETURN_TEXT` | Custom text to append to hello responses | (empty) |
| `COMPONENT` | Component name identifier | `component0` |

Example:
```bash
docker run -d -p 8080:8080 \
  -e RETURN_TEXT="from container" \
  -e COMPONENT="service-1" \
  --name multitoolserver \
  przemekmalak/multitoolserver
```

## ☁️ Google Cloud Run

### Deploy to Cloud Run

```bash
# Build and push to Docker Hub (for AMD64 architecture)
docker buildx build --platform linux/amd64 -t przemekmalak/multitoolserver:latest --push .

# Deploy to Cloud Run
gcloud run deploy multitoolserver \
  --image przemekmalak/multitoolserver:latest \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --port 8080 \
  --set-env-vars RETURN_TEXT="from cloud run"

# Get the service URL
gcloud run services describe multitoolserver --region us-central1 --format 'value(status.url)'
```

## ☸️ Kubernetes

### Quick Deployment

```bash
# Deploy using kubectl
kubectl run multitool \
  --image=przemekmalak/multitoolserver \
  --env="RETURN_TEXT=service 1" \
  --port=8080

# Expose the service
kubectl expose deployment multitool --port=80 --target-port=8080
```

### Using YAML Files

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: multitool
spec:
  replicas: 1
  selector:
    matchLabels:
      app: multitool
  template:
    metadata:
      labels:
        app: multitool
    spec:
      containers:
      - name: multitool
        image: przemekmalak/multitoolserver
        ports:
        - containerPort: 8080
        env:
        - name: RETURN_TEXT
          value: "from kubernetes"
```

## 🛠️ Development

### Prerequisites

- Go 1.23 or later
- Docker (optional)

### Local Development

```bash
# Navigate to source directory
cd src

# Build the application
go build -o multitoolserver .

# Run locally
./multitoolserver

# Or run directly with go
go run serve.go
```

### Testing

```bash
# Test basic functionality
curl http://localhost:8080/hello

# Test path traversal protection
curl "http://localhost:8080/ls?path=../../../etc"
# Should return: "Invalid path: path traversal not allowed"

# Test external request
curl "http://localhost:8080/req?url=https://httpbin.org/get"

# Test error endpoints
curl http://localhost:8080/error
curl http://localhost:8080/error2
```

## 📊 Example Responses

### Hello Endpoint
```bash
$ curl http://localhost:8080/hello
HostName: multitoolserver IP Address: 172.17.0.2
```

### Directory Listing
```bash
$ curl "http://localhost:8080/ls?path=/tmp"
Directory: /tmp
file1.txt
file2.txt
```

### Environment Variables
```bash
$ curl http://localhost:8080/env
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
HOSTNAME=multitoolserver
...
```

## 🔧 Configuration

The server runs on port 8080 by default. To change the port, modify the `serve.go` file:

```go
http.ListenAndServe(":8080", nil)  // Change 8080 to your desired port
```

## 🚨 Security Considerations

- The `/req` endpoint can make requests to any URL, but only valid `http` and `https` URLs are allowed. Use with caution in production.
- The `/ls` endpoint can access any directory on the filesystem. Consider restricting access in production.
- The server is designed for testing and debugging, not for production use without additional security measures.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is open source and available under the [MIT License](LICENSE).

## 🆘 Support

For issues and questions:
- Create an issue in the repository
- Check the existing documentation
- Review the code examples above

---

**Note**: This server is designed for testing and debugging purposes. Use appropriate security measures when deploying in production environments.
