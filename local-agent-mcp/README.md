# bbs-go MCP

`bbs-go-mcp` is a local Python MCP server for bbs-go article and article-comment workflows. It runs over stdio and stores the local `bbsgo_token` credential outside the Agent context.

## Install

```bash
pipx install .
bbs-go-mcp config init
bbs-go-mcp auth import-cookie
bbs-go-mcp run --stdio
```

The default target is `http://127.0.0.1:8082`. Update `config.json` before connecting to another bbs-go instance.

## Safety

- Write tools require `confirm=true`.
- `articleUrl` is restricted to the configured bbs-go host.
- Logs and tool output never include Cookie, Token, passwords, or full article content.
- `bbs_logout` revokes the remote session before deleting the local credential.

See `docs/mcp-client-config.md` and `docs/security.md`.
