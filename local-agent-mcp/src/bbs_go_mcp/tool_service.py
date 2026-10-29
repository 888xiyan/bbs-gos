from __future__ import annotations

import json
import time
from collections import deque
from collections.abc import Callable
from dataclasses import dataclass
from threading import Lock
from typing import Any
from uuid import uuid4

from .api_client import BbsApiClient
from .category_resolver import resolve_category
from .config import Config
from .credential_store import CredentialStore
from .errors import BbsMcpError, ErrorCode, error_result, validation_error
from .security import resolve_article_id, validate_content


@dataclass
class ToolResult:
    value: dict[str, Any]


class RateLimiter:
    def __init__(self, max_calls: int = 30, window_seconds: int = 60) -> None:
        self.max_calls = max_calls
        self.window_seconds = window_seconds
        self.calls: deque[float] = deque()
        self.lock = Lock()

    def check(self) -> None:
        now = time.monotonic()
        with self.lock:
            while self.calls and now - self.calls[0] > self.window_seconds:
                self.calls.popleft()
            if len(self.calls) >= self.max_calls:
                raise BbsMcpError(ErrorCode.RATE_LIMITED, "local MCP rate limit exceeded", True)
            self.calls.append(now)


class ToolService:
    def __init__(
        self,
        config: Config,
        credentials: CredentialStore,
        client_factory: Callable[..., BbsApiClient] = BbsApiClient,
    ) -> None:
        self.config = config
        self.credentials = credentials
        self.client_factory = client_factory
        self.rate_limiter = RateLimiter()

    def _client(self, request_id: str) -> BbsApiClient:
        credential = self.credentials.load()
        if credential is None:
            raise BbsMcpError(ErrorCode.AUTH_REQUIRED, "local credential is missing", False, request_id)
        return self.client_factory(self.config, credential)

    @staticmethod
    def _request_id() -> str:
        return str(uuid4())

    def _confirm(self, confirm: bool, preview: dict[str, Any], request_id: str) -> dict[str, Any]:
        if confirm:
            return preview
        return {
            "success": False,
            "error": ErrorCode.CONFIRMATION_REQUIRED.value,
            "message": "set confirm=true to execute this write",
            "retryable": False,
            "requestId": request_id,
            "requiredConfirmation": True,
            "ready": True,
            **preview,
        }

    def _call(self, request_id: str, operation: Any) -> dict[str, Any]:
        self.rate_limiter.check()
        try:
            with self._client(request_id) as client:
                return {"success": True, "requestId": request_id, "data": operation(client)}
        except Exception as exc:
            return error_result(exc, request_id)

    def status(self) -> dict[str, Any]:
        request_id = self._request_id()
        try:
            with self._client(request_id) as client:
                return {
                    "success": True,
                    "requestId": request_id,
                    "data": {
                        "site": client.get_site_config(request_id),
                        "user": client.get_current_user(request_id),
                        "authenticated": True,
                    },
                }
        except Exception as exc:
            return error_result(exc, request_id)

    def auth_status(self) -> dict[str, Any]:
        request_id = self._request_id()
        try:
            with self._client(request_id) as client:
                user = client.get_current_user(request_id)
                return {
                    "success": True,
                    "requestId": request_id,
                    "data": {
                        "authenticated": bool(user),
                        "user": {"id": user.get("id"), "nickname": user.get("nickname")} if user else None,
                    },
                }
        except Exception as exc:
            return error_result(exc, request_id)

    def list_categories(self, include_disabled: bool = False) -> dict[str, Any]:
        request_id = self._request_id()

        def operation(client: BbsApiClient) -> Any:
            categories = client.list_categories(request_id)
            return categories if include_disabled else _filter_categories(categories)

        return self._call(request_id, operation)

    def get_article(self, article_id: int | str | None = None, article_url: str | None = None) -> dict[str, Any]:
        request_id = self._request_id()
        try:
            resolved_id = resolve_article_id(article_id, article_url, self.config)
            with self._client(request_id) as client:
                return {"success": True, "requestId": request_id, "data": client.get_article(resolved_id, request_id)}
        except Exception as exc:
            return error_result(exc, request_id)

    def list_my_articles(self, cursor: str = "") -> dict[str, Any]:
        request_id = self._request_id()

        def operation(client: BbsApiClient) -> Any:
            user = client.get_current_user(request_id)
            if not user:
                raise BbsMcpError(ErrorCode.AUTH_REQUIRED, "current user is unavailable", False, request_id)
            return client.list_my_articles(user.get("id"), cursor, request_id)

        return self._call(request_id, operation)

    def list_comments(self, article_id: int | str | None = None, article_url: str | None = None, cursor: str = "") -> dict[str, Any]:
        request_id = self._request_id()
        try:
            resolved_id = resolve_article_id(article_id, article_url, self.config)
            with self._client(request_id) as client:
                return {
                    "success": True,
                    "requestId": request_id,
                    "data": client.list_comments(resolved_id, cursor, request_id),
                }
        except Exception as exc:
            return error_result(exc, request_id)

    def publish_article(
        self,
        title: str,
        content: str,
        category_id: int | str | None = None,
        category_path: str | None = None,
        tags: list[str] | None = None,
        summary: str = "",
        content_type: str = "markdown",
        allow_comments: bool = True,
        confirm: bool = False,
    ) -> dict[str, Any]:
        request_id = self._request_id()
        try:
            with self._client(request_id) as client:
                category = resolve_category(client.list_categories(request_id), category_id, category_path)
                title_value = validate_content(title, self.config, "title")
                content_value = validate_content(content, self.config, "content")
                body = {
                    "title": title_value,
                    "categoryId": int(category["id"]),
                    "summary": summary.strip(),
                    "content": content_value,
                    "contentType": content_type,
                    "tags": ",".join(tags or []),
                    "allowComments": bool(allow_comments),
                }
                preview = {
                    "resolvedCategory": {
                        "id": category.get("id"),
                        "path": category.get("path"),
                        "urlPath": category.get("urlPath"),
                    },
                    "warnings": [],
                }
                if not confirm:
                    return self._confirm(False, preview, request_id)
                return {"success": True, "requestId": request_id, "data": client.publish_article(body, request_id)}
        except Exception as exc:
            return error_result(exc, request_id)

    def update_article(
        self,
        article_id: int | str,
        title: str | None = None,
        content: str | None = None,
        category_id: int | str | None = None,
        category_path: str | None = None,
        tags: list[str] | None = None,
        summary: str | None = None,
        allow_comments: bool | None = None,
        confirm: bool = False,
    ) -> dict[str, Any]:
        request_id = self._request_id()
        try:
            resolved_id = resolve_article_id(article_id, None, self.config)
            with self._client(request_id) as client:
                edit = client.get_article_edit(resolved_id, request_id)
                category = resolve_category(client.list_categories(request_id), category_id or edit.get("categoryId"), category_path)
                body = {
                    "title": title.strip() if title is not None else edit.get("title", ""),
                    "categoryId": int(category["id"]),
                    "summary": summary.strip() if summary is not None else "",
                    "content": validate_content(content, self.config) if content is not None else edit.get("content", ""),
                    "tags": ",".join(tags if tags is not None else edit.get("tags") or []),
                    "cover": json.dumps(edit.get("cover")) if edit.get("cover") else "",
                    "allowComments": bool(allow_comments) if allow_comments is not None else bool(edit.get("allowComments", True)),
                }
                if not confirm:
                    return self._confirm(False, {"resolvedCategory": category, "warnings": []}, request_id)
                return {"success": True, "requestId": request_id, "data": client.update_article(resolved_id, body, request_id)}
        except Exception as exc:
            return error_result(exc, request_id)

    def set_comment_policy(self, article_id: int | str, allow_comments: bool, confirm: bool = False) -> dict[str, Any]:
        return self.update_article(article_id, allow_comments=allow_comments, confirm=confirm)

    def publish_comment(
        self,
        content: str,
        article_id: int | str | None = None,
        article_url: str | None = None,
        reply_to_comment_id: int | str | None = None,
        confirm: bool = False,
    ) -> dict[str, Any]:
        request_id = self._request_id()
        try:
            resolved_id = resolve_article_id(article_id, article_url, self.config)
            content_value = validate_content(content, self.config, "content")
            with self._client(request_id) as client:
                article = client.get_article(resolved_id, request_id)
                if article.get("allowComments") is False:
                    raise BbsMcpError(ErrorCode.COMMENTS_DISABLED, "comments are disabled for this article", False, request_id)
                entity_type = "article"
                entity_id = str(resolved_id)
                if reply_to_comment_id not in (None, ""):
                    parent_id = int(reply_to_comment_id)
                    comments = client.list_comments(resolved_id, "", request_id).get("results") or []
                    if not _comment_belongs_to_article(comments, parent_id):
                        raise BbsMcpError(ErrorCode.COMMENT_NOT_FOUND, "reply comment does not belong to article", False, request_id)
                    entity_type = "comment"
                    entity_id = str(parent_id)
                body = {
                    "entityType": entity_type,
                    "entityId": entity_id,
                    "content": content_value,
                }
                if not confirm:
                    return self._confirm(False, {"articleId": resolved_id, "warnings": []}, request_id)
                return {"success": True, "requestId": request_id, "data": client.publish_comment(body, request_id)}
        except Exception as exc:
            return error_result(exc, request_id)

    def delete_comment(self, comment_id: int | str, confirm: bool = False) -> dict[str, Any]:
        request_id = self._request_id()
        try:
            comment_id_value = int(comment_id)
            if comment_id_value <= 0:
                raise validation_error("commentId must be positive")
            if not confirm:
                return self._confirm(False, {"commentId": comment_id_value, "warnings": []}, request_id)
            with self._client(request_id) as client:
                return {"success": True, "requestId": request_id, "data": client.delete_comment(comment_id_value, request_id)}
        except Exception as exc:
            return error_result(exc, request_id)

    def logout(self, confirm: bool = False) -> dict[str, Any]:
        request_id = self._request_id()
        try:
            if not confirm:
                return self._confirm(False, {"warnings": []}, request_id)
            with self._client(request_id) as client:
                result = client.logout(request_id)
            self.credentials.delete()
            return {"success": True, "requestId": request_id, "data": result}
        except Exception as exc:
            return error_result(exc, request_id)


def _comment_belongs_to_article(comments: list[dict[str, Any]], comment_id: int) -> bool:
    for comment in comments:
        if int(comment.get("id", 0)) == comment_id:
            return True
        replies = comment.get("replies")
        if isinstance(replies, dict) and _comment_belongs_to_article(replies.get("results") or [], comment_id):
            return True
        if isinstance(replies, list) and _comment_belongs_to_article(replies, comment_id):
            return True
    return False


def _filter_categories(nodes: list[dict[str, Any]]) -> list[dict[str, Any]]:
    result: list[dict[str, Any]] = []
    for node in nodes:
        children = _filter_categories(node.get("children") or [])
        if node.get("status", 0) == 0:
            item = dict(node)
            item["children"] = children
            result.append(item)
    return result
