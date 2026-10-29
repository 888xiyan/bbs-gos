from __future__ import annotations

import getpass

from .api_client import BbsApiClient
from .config import Config
from .credential_store import Credential, CredentialStore
from .errors import BbsMcpError, ErrorCode, validation_error


def normalize_cookie(value: str) -> str:
    raw = value.strip()
    if not raw:
        raise validation_error("cookie is required")
    for prefix in ("Cookie:", "cookie:"):
        if raw.startswith(prefix):
            raw = raw[len(prefix) :].strip()
    if raw.startswith("bbsgo_token="):
        raw = raw.split(";", 1)[0][len("bbsgo_token=") :]
    if not raw or len(raw) > 512 or any(char in raw for char in "\r\n"):
        raise validation_error("invalid cookie value")
    return raw


def import_cookie(config: Config, store: CredentialStore, value: str | None = None) -> Credential:
    raw = value if value is not None else getpass.getpass("bbsgo_token cookie: ")
    cookie = normalize_cookie(raw)
    with BbsApiClient(config, Credential(cookie=cookie)) as client:
        user = client.get_current_user()
        if not user:
            raise BbsMcpError(ErrorCode.AUTH_REQUIRED, "cookie did not resolve to a current user")
    return store.save(cookie)


def auth_status(config: Config, store: CredentialStore) -> dict[str, object]:
    credential = store.load()
    if credential is None:
        return {"authenticated": False, "user": None}
    try:
        with BbsApiClient(config, credential) as client:
            user = client.get_current_user()
            return {
                "authenticated": bool(user),
                "user": {"id": user.get("id"), "nickname": user.get("nickname")} if user else None,
            }
    except BbsMcpError as exc:
        return {"authenticated": False, "error": exc.code.value}


def logout(config: Config, store: CredentialStore) -> None:
    credential = store.load()
    if credential is None:
        return
    try:
        with BbsApiClient(config, credential) as client:
            client.logout()
    finally:
        store.delete()
