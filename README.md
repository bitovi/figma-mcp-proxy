# Figma MCP Proxy

A Go-based HTTP proxy server that enhances Model Context Protocol (MCP) tool calls by automatically adding Figma file parameters and opening Figma designs when processing tool calls.

- For more information, check out the [Cascading: Cloud AI Implements Figma and Jira Usage Guide](https://bitovi.atlassian.net/wiki/spaces/AIEnabledDevelopment/pages/1517289538/Cascading+v2+Cloud+AI+implements+Figma+and+Jira).
- Need help? Find Bitovi on [Discord](https://discord.gg/J7ejFsZnJ4) or [hire us](https://www.bitovi.com/services/ai-consulting).

👉 Bitovi can help you integrate this into your own SDLC workflow: [AI for Software Teams](https://www.bitovi.com/ai-for-software-teams)


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

## Restart the service

### Recommended: Use the restart script
1. Remote desktop into the Windows server
2. Open PowerShell and run:
```powershell
cd C:\figma-mcp-proxy
& ".\restart.ps1"
```

This will stop the task, pull latest code, rebuild, and restart.

### Manual restart (without code update)
If you just need to restart without pulling new code:
```powershell
Stop-ScheduledTask -TaskName 'FigmaProxy' -ErrorAction SilentlyContinue
Start-ScheduledTask -TaskName 'FigmaProxy'
```

### Check task status
```powershell
Get-ScheduledTask -TaskName 'FigmaProxy' | Get-ScheduledTaskInfo
```

### View logs
```powershell
Get-Content C:\figma-mcp-proxy\logs\proxy.log -Tail 50
```

## Usage

### Starting the proxy

```bash
go run main.go
```

Or with custom configuration:

```bash
TARGET_URL=http://localhost:3000 PORT=8080 go run main.go
```
