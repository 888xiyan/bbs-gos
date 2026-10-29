from __future__ import annotations

import json
from dataclasses import asdict, dataclass, field
from pathlib import Path
from typing import Any
from urllib.parse import urlparse

from platformdirs import user_config_dir

from .errors import validation_error


@dataclass
class Config:
    baseUrl: str = "http://127.0.0.1:8082"
    timeoutSeconds: int = 30
    defaultConfirm: bool = False
    allowedArticleUrlHosts: list[str] = field(default_factory=list)
    maxContentLength: int = 200000

    @property
    def allowed_hosts(self) -> set[str]:
        return set(self.allowedArticleUrlHosts) or {urlparse(self.baseUrl).netloc.lower()}


def config_dir() -> Path:
    return Path(user_config_dir("bbs-go-mcp"))


def config_path() -> Path:
    return config_dir() / "config.json"


def validate_config(config: Config) -> Config:
    parsed = urlparse(config.baseUrl)
    if parsed.scheme not in {"http", "https"} or not parsed.netloc:
        raise validation_error("baseUrl must be an absolute http or https URL")
    if not 1 <= config.timeoutSeconds <= 120:
        raise validation_error("timeoutSeconds must be between 1 and 120")
    if not 1 <= config.maxContentLength <= 2_000_000:
        raise validation_error("maxContentLength must be between 1 and 2000000")
    config.allowedArticleUrlHosts = [host.lower() for host in config.allowedArticleUrlHosts if host]
    return config


def load_config(path: Path | None = None) -> Config:
    target = path or config_path()
    if not target.exists():
        return validate_config(Config())
    try:
        data: dict[str, Any] = json.loads(target.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        raise validation_error(f"invalid config file: {exc}") from exc
    return validate_config(Config(**data))


def save_config(config: Config, path: Path | None = None) -> Path:
    target = path or config_path()
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(json.dumps(asdict(validate_config(config)), indent=2) + "\n", encoding="utf-8")
    target.chmod(0o600)
    return target
