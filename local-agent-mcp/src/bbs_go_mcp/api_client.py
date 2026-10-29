from __future__ import annotations

from dataclasses import dataclass
from typing import Any

import httpx

from .config import Config
from .credential_store import Credential
from .errors import BbsMcpError, ErrorCode


@dataclass
class ApiResponse:
    data: Any
    request_id: str


class BbsApiClient:
    def __init__(self, config: Config, credential: Credential) -> None:
        self.config = config
        self.credential = credential
        self.client = httpx.Client(
            base_url=config.baseUrl,
            timeout=config.timeoutSeconds,
            follow_redirects=False,
            headers={"Accept": "application/json"},
            cookies={"bbsgo_token": credential.cookie},
        )

    def close(self) -> None:
        self.client.close()

    def __enter__(self) -> BbsApiClient:
        return self

    def __exit__(self, *_: object) -> None:
        self.close()

    def request(
        self,
        method: str,
        path: str,
        *,
        json_body: dict[str, Any] | None = None,
        params: dict[str, Any] | None = None,
        request_id: str = "",
    ) -> ApiResponse:
        try:
            response = self.client.request(method, path, json=json_body, params=params)
        except httpx.TimeoutException as exc:
            raise BbsMcpError(ErrorCode.NETWORK_ERROR, "request timed out", True, request_id) from exc
        except httpx.HTTPError as exc:
            raise BbsMcpError(ErrorCode.NETWORK_ERROR, "network request failed", True, request_id) from exc

        try:
            payload = response.json()
        except ValueError:
            payload = {}

        if response.status_code == 401:
            raise BbsMcpError(ErrorCode.AUTH_REQUIRED, "authentication required", False, request_id)
        if response.status_code == 403:
            raise BbsMcpError(ErrorCode.PERMISSION_DENIED, "permission denied", False, request_id)
        if response.status_code == 429:
            raise BbsMcpError(ErrorCode.RATE_LIMITED, "rate limit exceeded", True, request_id)
        if response.status_code >= 500:
            raise BbsMcpError(ErrorCode.NETWORK_ERROR, "bbs-go server error", True, request_id)

        if isinstance(payload, dict) and payload.get("success") is False:
            message = str(payload.get("message") or "request failed")
            code = self._map_message(message, response.status_code)
            raise BbsMcpError(code, message, code in {ErrorCode.RATE_LIMITED, ErrorCode.NETWORK_ERROR}, request_id)
        if response.status_code >= 400:
            message = str(payload.get("message") or response.text or "request failed") if isinstance(payload, dict) else "request failed"
            raise BbsMcpError(self._map_message(message, response.status_code), message, False, request_id)

        data = payload.get("data") if isinstance(payload, dict) else payload
        return ApiResponse(data=data, request_id=request_id)

    @staticmethod
    def _map_message(message: str, status: int) -> ErrorCode:
        lower = message.lower()
        if "comment" in lower and ("disabled" in lower or "closed" in lower):
            return ErrorCode.COMMENTS_DISABLED
        if "category" in lower and "found" in lower:
            return ErrorCode.CATEGORY_NOT_FOUND
        if status == 404 or "not found" in lower:
            return ErrorCode.ARTICLE_NOT_FOUND
        if "permission" in lower or "forbidden" in lower:
            return ErrorCode.PERMISSION_DENIED
        if "login" in lower or "auth" in lower:
            return ErrorCode.AUTH_REQUIRED
        if "invalid" in lower or "required" in lower:
            return ErrorCode.VALIDATION_ERROR
        return ErrorCode.VALIDATION_ERROR

    def get_site_config(self, request_id: str = "") -> Any:
        return self.request("GET", "/api/config/configs", request_id=request_id).data

    def get_current_user(self, request_id: str = "") -> Any:
        return self.request("GET", "/api/user/current", request_id=request_id).data

    def list_categories(self, request_id: str = "") -> Any:
        return self.request("GET", "/api/category/tree", request_id=request_id).data

    def get_article(self, article_id: int, request_id: str = "") -> Any:
        return self.request("GET", f"/api/article/{article_id}", request_id=request_id).data

    def get_article_edit(self, article_id: int, request_id: str = "") -> Any:
        return self.request("GET", f"/api/article/edit/{article_id}", request_id=request_id).data

    def list_my_articles(self, user_id: str | int, cursor: str = "", request_id: str = "") -> Any:
        return self.request(
            "GET",
            "/api/article/user_articles",
            params={"userId": user_id, "cursor": cursor},
            request_id=request_id,
        ).data

    def list_comments(self, article_id: int, cursor: str = "", request_id: str = "") -> Any:
        return self.request(
            "GET",
            "/api/comment/comments",
            params={"entityType": "article", "entityId": article_id, "cursor": cursor},
            request_id=request_id,
        ).data

    def publish_article(self, body: dict[str, Any], request_id: str = "") -> Any:
        return self.request("POST", "/api/article/create", json_body=body, request_id=request_id).data

    def update_article(self, article_id: int, body: dict[str, Any], request_id: str = "") -> Any:
        return self.request("POST", f"/api/article/edit/{article_id}", json_body=body, request_id=request_id).data

    def publish_comment(self, body: dict[str, Any], request_id: str = "") -> Any:
        return self.request("POST", "/api/comment/create", json_body=body, request_id=request_id).data

    def delete_comment(self, comment_id: int, request_id: str = "") -> Any:
        return self.request("POST", f"/api/comment/delete/{comment_id}", request_id=request_id).data

    def logout(self, request_id: str = "") -> Any:
        return self.request("GET", "/api/login/signout", request_id=request_id).data
