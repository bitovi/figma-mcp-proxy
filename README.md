# Figma MCP Proxy

A Go-based HTTP proxy server that enhances Model Context Protocol (MCP) tool calls by automatically adding Figma file parameters and opening Figma designs when processing tool calls.

## What it does

This proxy acts as an intermediary between MCP clients and MCP servers, providing the following functionality:

### 1. Tool Schema Enhancement

When a client requests the list of available tools (`tools/list`), the proxy:

- Identifies tools that have a `nodeId` parameter in their input schema
- Automatically adds two additional parameters to these tools:
  - `fileKey`: The Figma file identifier (extracted from URLs like `https://figma.com/design/1234/5678?node-id=1-2`)
  - `fileName`: The Figma file name (extracted from the same URL format)
- Updates tool descriptions to explain how to extract these parameters from Figma URLs

### 2. Automatic Figma Design Opening

When processing tool calls that include both `fileKey` and `fileName` parameters, the proxy:

- Automatically opens the specified Figma design using the system's default Figma application
- Uses the `figma://` URL scheme to launch directly to the design
- Supports macOS, Windows, and Linux operating systems
- Adds a 2-second delay to allow Figma to fully launch before proceeding

### 3. Request Logging

The proxy logs all MCP requests for debugging and monitoring purposes, including:

- HTTP method, URL path, and remote address
- Request body content (for requests under 1MB)

## Architecture

The proxy uses Go's `httputil.ReverseProxy` to:

- **Director Function**: Intercepts incoming requests, parses MCP payloads, and triggers Figma design opening when appropriate
- **ModifyResponse Function**: Modifies `tools/list` responses to add the additional file parameters
- **Health Endpoint**: Provides a `/health` endpoint for monitoring proxy status

## Configuration

The proxy can be configured using environment variables:

- `TARGET_URL`: The MCP server to proxy requests to (default: `http://localhost:3845`)
- `PORT`: The port to run the proxy server on (default: `3846`)

## Usage

### Starting the proxy

```bash
go run main.go
```

Or with custom configuration:

```bash
TARGET_URL=http://localhost:3000 PORT=8080 go run main.go
```

### Example Tool Call Flow

1. **Client requests tools**: `POST /mcp` with `{"method": "tools/list"}`
2. **Proxy enhances response**: Adds `fileKey` and `fileName` parameters to relevant tools
3. **Client calls tool**: `POST /mcp` with tool call including `fileKey` and `fileName`
4. **Proxy opens Figma**: Automatically launches `figma://design/{fileKey}/{fileName}`
5. **Request forwarded**: Original tool call is forwarded to the target MCP server

### URL Parameter Extraction

For Figma URLs like `https://figma.com/design/JqWii6wYby2bPqnaaALroQ/USER-10?node-id=1-119`:

- `fileKey`: `JqWii6wYby2bPqnaaALroQ`
- `fileName`: `USER-10`

## Health Check

The proxy provides a health check endpoint at `/health` that returns:

```json
{
  "status": "OK",
  "targetURL": "http://localhost:3845"
}
```

## Requirements

- Go 1.21 or later
- Figma desktop application installed on the system
- Access to the target MCP server

## Cross-Platform Support

The proxy supports opening Figma designs on:

- **macOS**: Uses `open figma://design/{fileKey}/{fileName}`
- **Windows**: Uses PowerShell `Start-Process figma://design/{fileKey}/{fileName}`
- **Linux**: Uses `xdg-open figma://design/{fileKey}/{fileName}`
