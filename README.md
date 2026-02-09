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

### 2. Automatic Figma Design Opening with Smart Retry

When processing tool calls that include both `fileKey` and `fileName` parameters, the proxy uses an optimistic approach:

- **First attempt**: Forwards the request immediately without opening any files (instant response if file already open)
- **Error detection**: Monitors response for errors indicating the file isn't open ("No node could be found", "AppStateTsApi is null", etc.)
- **Automatic retry**: If file-not-open error detected, opens the design and retries the request automatically
- Uses the `figma://` URL scheme to launch directly to the design
- Supports macOS, Windows, and Linux operating systems
- Adds a configurable delay (default 30 seconds) when opening files to ensure full load
- **Performance**: Zero delay for files already open, only pays the time cost when switching files

## Configuration

The proxy can be configured using environment variables:

- `TARGET_URL`: The MCP server to proxy requests to (default: `http://localhost:3845`)
- `PORT`: The port to run the proxy server on (default: `3846`)
- `FIGMA_OPEN_DELAY_SECONDS`: Time to wait after opening a Figma file before proceeding with the tool call (default: `30` seconds). Increase this for slow remote VM scenarios where files take longer to load.

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

## Troubleshooting

### Figma MCP tools fail after disconnecting RDP

**Problem**: Some Figma MCP tools (particularly `get_design_context`) stop working after disconnecting from the RDP session, even though Figma remains running.

**Cause**: When you close RDP normally, Windows marks the session as "Disconnected", which changes how the desktop/graphics context is maintained. Figma continues running, but tools requiring an active graphics session fail.

#### Solution: Automated Disconnect Script (Recommended)

Instead of closing the RDP window normally, run this script before disconnecting:

```powershell
cd C:\figma-mcp-proxy
& ".\disconnect-rdp.ps1"
```

This automatically transfers your session to the console, keeping it active.

#### Manual Method

If the script fails, you can manually transfer the session:

1. Open Command Prompt as Administrator
2. Run: `query session`
3. Find your session ID (State = Active)
4. Run: `tscon <ID> /dest:console` (e.g., `tscon 2 /dest:console`)

## Usage

### Starting the proxy

```bash
go run main.go
```

Or with custom configuration:

```bash
TARGET_URL=http://localhost:3000 PORT=8080 go run main.go
```
