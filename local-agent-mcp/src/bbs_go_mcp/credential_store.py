from __future__ import annotations

import json
import os
import secrets
import subprocess
from dataclasses import dataclass
from pathlib import Path

from cryptography.hazmat.primitives.ciphers.aead import AESGCM

from .config import config_dir
from .errors import BbsMcpError, ErrorCode
from .security import atomic_write

SERVICE_NAME = "bbs-go-mcp"
COOKIE_KEY = "bbsgo_token"


@dataclass(frozen=True)
class Credential:
    cookie: str
    username: str | None = None


class CredentialStore:
    def __init__(self, directory: Path | None = None, use_keyring: bool = True) -> None:
        self.directory = directory or config_dir()
        self.use_keyring = use_keyring
        self.credentials_path = self.directory / "credentials.enc"
        self.master_key_path = self.directory / "master.key"

    def load(self) -> Credential | None:
        if self.use_keyring:
            credential = self._load_keyring()
            if credential:
                return credential
        return self._load_file()

    def save(self, cookie: str) -> Credential:
        if not cookie.strip():
            raise BbsMcpError(ErrorCode.VALIDATION_ERROR, "cookie is required")
        credential = Credential(cookie=cookie.strip())
        if self.use_keyring:
            try:
                self._save_keyring(credential)
                return credential
            except Exception:
                pass
        self._save_file(credential)
        return credential

    def delete(self) -> None:
        if self.use_keyring:
            try:
                self._delete_keyring()
            except Exception:
                pass
        for path in (self.credentials_path, self.master_key_path):
            if path.exists():
                path.unlink()

    def _load_keyring(self) -> Credential | None:
        try:
            import keyring

            cookie = keyring.get_password(SERVICE_NAME, COOKIE_KEY)
            return Credential(cookie=cookie) if cookie else None
        except Exception:
            return None

    def _save_keyring(self, credential: Credential) -> None:
        import keyring

        keyring.set_password(SERVICE_NAME, COOKIE_KEY, credential.cookie)

    def _delete_keyring(self) -> None:
        import keyring

        keyring.delete_password(SERVICE_NAME, COOKIE_KEY)

    def _load_file(self) -> Credential | None:
        if not self.credentials_path.exists():
            return None
        try:
            key = self.master_key_path.read_bytes()
            encrypted = self.credentials_path.read_bytes()
            plaintext = AESGCM(key).decrypt(encrypted[:12], encrypted[12:], None)
            data = json.loads(plaintext.decode("utf-8"))
            return Credential(cookie=data["cookie"], username=data.get("username"))
        except Exception as exc:
            raise BbsMcpError(ErrorCode.STORAGE_UNAVAILABLE, "credential file cannot be decrypted") from exc

    def _save_file(self, credential: Credential) -> None:
        self.directory.mkdir(parents=True, exist_ok=True)
        if not self.master_key_path.exists():
            atomic_write(self.master_key_path, secrets.token_bytes(32), 0o600)
        key = self.master_key_path.read_bytes()
        plaintext = json.dumps({"cookie": credential.cookie}, separators=(",", ":")).encode()
        nonce = secrets.token_bytes(12)
        atomic_write(self.credentials_path, nonce + AESGCM(key).encrypt(nonce, plaintext, None), 0o600)
        self._secure_file(self.credentials_path)
        self._secure_file(self.master_key_path)

    def _secure_file(self, path: Path) -> None:
        if os.name == "nt":
            user = os.getenv("USERNAME") or os.getenv("USER") or ""
            if user:
                subprocess.run(
                    ["icacls", str(path), "/inheritance:r", "/grant:r", f"{user}:F"],
                    check=False,
                    capture_output=True,
                )
        else:
            path.chmod(0o600)
