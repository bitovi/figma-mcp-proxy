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



