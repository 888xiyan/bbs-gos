from __future__ import annotations

import argparse
import json
import sys

from .auth import auth_status, import_cookie, logout
from .config import Config, config_path, load_config, save_config
from .credential_store import CredentialStore
from .errors import validation_error
from .server import run_stdio


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="bbs-go-mcp")
    subparsers = parser.add_subparsers(dest="command", required=True)

    run = subparsers.add_parser("run", help="run the MCP server")
    run.add_argument("--stdio", action="store_true", help="use MCP stdio transport")

    config = subparsers.add_parser("config", help="manage local configuration")
    config_subparsers = config.add_subparsers(dest="config_command", required=True)
    config_subparsers.add_parser("init", help="write a default configuration file")

    auth = subparsers.add_parser("auth", help="manage local credentials")
    auth_subparsers = auth.add_subparsers(dest="auth_command", required=True)
    auth_subparsers.add_parser("import-cookie", help="import a bbsgo_token cookie")
    auth_subparsers.add_parser("status", help="show authentication status")
    auth_subparsers.add_parser("logout", help="revoke the remote session and delete local credentials")

    subparsers.add_parser("doctor", help="check local configuration and credential storage")
    return parser


def main() -> None:
    args = build_parser().parse_args()
    try:
        config = load_config()
        if args.command == "run":
            if not args.stdio:
                raise validation_error("run requires --stdio")
            run_stdio(config)
            return
        if args.command == "config" and args.config_command == "init":
            path = save_config(Config())
            print(path)
            return
        store = CredentialStore()
        if args.command == "auth":
            if args.auth_command == "import-cookie":
                credential = import_cookie(config, store)
                print(json.dumps({"authenticated": True, "username": credential.username}, ensure_ascii=False))
                return
            if args.auth_command == "status":
                print(json.dumps(auth_status(config, store), ensure_ascii=False))
                return
            if args.auth_command == "logout":
                logout(config, store)
                print(json.dumps({"success": True}))
                return
        if args.command == "doctor":
            loaded_credential = store.load()
            print(
                json.dumps(
                    {
                        "python": sys.version.split()[0],
                        "config": str(config_path()),
                        "baseUrl": config.baseUrl,
                        "credentialStored": loaded_credential is not None,
                        "credentialMode": "keyring-or-encrypted-file",
                    },
                    ensure_ascii=False,
                )
            )
            return
        raise ValueError("unknown command")
    except Exception as exc:
        message = str(exc) or exc.__class__.__name__
        print(json.dumps({"success": False, "message": message}, ensure_ascii=False), file=sys.stderr)
        raise SystemExit(2) from exc
