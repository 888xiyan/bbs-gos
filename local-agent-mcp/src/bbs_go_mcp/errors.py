from __future__ import annotations

from dataclasses import dataclass
from enum import StrEnum
from typing import Any
from uuid import uuid4


class ErrorCode(StrEnum):
    AUTH_REQUIRED = "AUTH_REQUIRED"
    PERMISSION_DENIED = "PERMISSION_DENIED"
    CATEGORY_NOT_FOUND = "CATEGORY_NOT_FOUND"
    CATEGORY_AMBIGUOUS = "CATEGORY_AMBIGUOUS"
    CATEGORY_NOT_POSTABLE = "CATEGORY_NOT_POSTABLE"
    COMMENTS_DISABLED = "COMMENTS_DISABLED"
    ARTICLE_NOT_FOUND = "ARTICLE_NOT_FOUND"
    COMMENT_NOT_FOUND = "COMMENT_NOT_FOUND"
    VALIDATION_ERROR = "VALIDATION_ERROR"
    RATE_LIMITED = "RATE_LIMITED"
    CONFIRMATION_REQUIRED = "CONFIRMATION_REQUIRED"
    NETWORK_ERROR = "NETWORK_ERROR"
    STORAGE_UNAVAILABLE = "STORAGE_UNAVAILABLE"


@dataclass
class BbsMcpError(Exception):
    code: ErrorCode
    message: str
    retryable: bool = False
    request_id: str = ""

    def __post_init__(self) -> None:
        if not self.request_id:
            self.request_id = str(uuid4())

    def __str__(self) -> str:
        return f"{self.code}: {self.message}"

    def result(self) -> dict[str, Any]:
        return {
            "success": False,
            "error": self.code.value,
            "message": self.message,
            "retryable": self.retryable,
            "requestId": self.request_id,
        }


def validation_error(message: str) -> BbsMcpError:
    return BbsMcpError(ErrorCode.VALIDATION_ERROR, message)


def error_result(error: Exception, request_id: str = "") -> dict[str, Any]:
    if isinstance(error, BbsMcpError):
        if request_id and not error.request_id:
            error.request_id = request_id
        return error.result()
    return BbsMcpError(ErrorCode.NETWORK_ERROR, str(error), True, request_id).result()
