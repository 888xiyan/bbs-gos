from __future__ import annotations

from typing import Any

from .errors import BbsMcpError, ErrorCode


def flatten_categories(nodes: list[dict[str, Any]] | None, parent_path: str = "") -> list[dict[str, Any]]:
    result: list[dict[str, Any]] = []
    for node in nodes or []:
        name = str(node.get("name") or "")
        path = str(node.get("path") or (f"{parent_path} / {name}" if parent_path else name))
        item = dict(node)
        item["path"] = path
        result.append(item)
        result.extend(flatten_categories(node.get("children") or [], path))
    return result


def resolve_category(
    categories: list[dict[str, Any]],
    category_id: int | str | None = None,
    category_path: str | None = None,
) -> dict[str, Any]:
    items = flatten_categories(categories)
    if category_id not in (None, ""):
        try:
            wanted_id = int(str(category_id))
        except ValueError as exc:
            raise BbsMcpError(ErrorCode.VALIDATION_ERROR, "categoryId must be a positive integer") from exc
        matches = [item for item in items if int(item.get("id", 0)) == wanted_id]
    else:
        wanted_path = normalize_category_path(category_path or "")
        if not wanted_path:
            raise BbsMcpError(ErrorCode.VALIDATION_ERROR, "categoryId or categoryPath is required")
        matches = [
            item
            for item in items
            if normalize_category_path(str(item.get("path") or "")) == wanted_path
        ]
    if not matches:
        raise BbsMcpError(ErrorCode.CATEGORY_NOT_FOUND, "category not found")
    if len(matches) > 1:
        raise BbsMcpError(ErrorCode.CATEGORY_AMBIGUOUS, "category path is ambiguous")
    category = matches[0]
    if category.get("allowPost") is False:
        raise BbsMcpError(ErrorCode.CATEGORY_NOT_POSTABLE, "category does not allow article publishing")
    return category


def normalize_category_path(path: str) -> str:
    return " / ".join(part.strip() for part in path.strip().split("/") if part.strip())
