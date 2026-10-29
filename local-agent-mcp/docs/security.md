# Security Notes

- The MCP server uses `bbsgo_token` only in its local HTTP client.
- The system keychain is preferred on macOS, Windows, and Linux.
- The fallback file uses AES-256-GCM with a separate `master.key` file.
- Unix credential files are written with mode `0600`; Windows files are restricted with an ACL.
- Tool output is limited to business results and stable error codes.
- URL inputs are restricted to the configured bbs-go host and reject query strings, fragments, redirects, and non-HTTP protocols.
- The final authorization and business checks remain in bbs-go.
