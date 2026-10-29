# MCP Client Configuration

Example stdio configuration:

```json
{
  "mcpServers": {
    "bbs-go": {
      "command": "bbs-go-mcp",
      "args": ["run", "--stdio"],
      "env": {
        "BBSGO_BASE_URL": "http://127.0.0.1:8082"
      }
    }
  }
}
```

Import and verify credentials before starting the MCP client:

```bash
bbs-go-mcp auth import-cookie
bbs-go-mcp auth status
```
