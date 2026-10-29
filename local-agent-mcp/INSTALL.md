# Install bbs-go MCP

The release archive contains the installable Python wheel and source distribution.

## Linux and macOS

```bash
python3 -m venv .venv
. .venv/bin/activate
python -m pip install bbs_go_mcp-0.1.0-py3-none-any.whl
bbs-go-mcp config init
bbs-go-mcp auth import-cookie
bbs-go-mcp run --stdio
```

## Windows PowerShell

```powershell
py -3 -m venv .venv
.venv\Scripts\Activate.ps1
python -m pip install bbs_go_mcp-0.1.0-py3-none-any.whl
bbs-go-mcp config init
bbs-go-mcp auth import-cookie
bbs-go-mcp run --stdio
```

The package targets Python 3.11 or later. The command runs a local stdio MCP server and does not open a public network port.
