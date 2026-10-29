from __future__ import annotations

import os
import re
import tempfile
from pathlib import Path
from urllib.parse import urlparse

from .config import Config
from .errors import validation_error

ARTICLE_URL_PATTERN = re.compile(r"^/article/(\d+)/?$")


def atomic_write(path: Path, data: bytes, mode: int = 0o600) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    descriptor, temp_name = tempfile.mkstemp(prefix=f".{path.name}.", dir=str(path.parent))
    temp_path = Path(temp_name)
    try:
        with os.fdopen(descriptor, "wb") as stream:
            stream.write(data)
            stream.flush()
            os.fsync(stream.fileno())
        os.chmod(temp_path, mode)
        os.replace(temp_path, path)
    finally:
        if temp_path.exists():
            temp_path.unlink()


def validate_content(content: str, config: Config, field_name: str = "content") -> str:
    value = content.strip()
    if not value:
        raise validation_error(f"{field_name} is required")
    if len(value.encode("utf-8")) > config.maxContentLength:
        raise validation_error(f"{field_name} exceeds maxContentLength")
    return value


def resolve_article_id(
    article_id: int | str | None,
    article_url: str | None,
    config: Config,
) -> int:
    if article_id not in (None, ""):
        try:
            value = int(str(article_id))
        except ValueError as exc:
            raise validation_error("articleId must be a positive integer") from exc
        if value <= 0:
            raise validation_error("articleId must be a positive integer")
        return value
    if not article_url:
        raise validation_error("articleId or articleUrl is required")
    parsed = urlparse(article_url)
    if parsed.scheme not in {"http", "https"}:
        raise validation_error("articleUrl must use http or https")
    if parsed.netloc.lower() not in config.allowed_hosts:
        raise validation_error("articleUrl host is not allowed")
    if parsed.query or parsed.fragment:
        raise validation_error("articleUrl must not contain query or fragment")
    match = ARTICLE_URL_PATTERN.match(parsed.path)
    if not match:
        raise validation_error("articleUrl must point to /article/:id")
    return int(match.group(1))
