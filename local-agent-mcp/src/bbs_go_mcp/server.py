from __future__ import annotations

from typing import Any

from mcp.server.mcpserver import MCPServer

from .config import Config
from .credential_store import CredentialStore
from .tool_service import ToolService


def create_server(
    config: Config | None = None,
    credentials: CredentialStore | None = None,
    service: ToolService | None = None,
) -> MCPServer:
    config = config or Config()
    credentials = credentials or CredentialStore()
    service = service or ToolService(config, credentials)
    mcp = MCPServer(
        name="bbs-go",
        instructions="Use the bbs-go article and article-comment tools. Write tools require confirm=true.",
    )

    @mcp.tool()
    def bbs_status() -> dict[str, Any]:
        """Return site status and the current local user without exposing credentials."""
        return service.status()

    @mcp.tool()
    def bbs_auth_status() -> dict[str, Any]:
        """Return local authentication status and basic user identity."""
        return service.auth_status()

    @mcp.tool()
    def bbs_list_categories(include_disabled: bool = False) -> dict[str, Any]:
        """List article categories with paths, URL paths, and publishing status."""
        return service.list_categories(include_disabled)

    @mcp.tool()
    def bbs_get_article(
        article_id: int | str | None = None,
        article_url: str | None = None,
    ) -> dict[str, Any]:
        """Read one article by numeric ID or an allowed same-origin article URL."""
        return service.get_article(article_id, article_url)

    @mcp.tool()
    def bbs_list_my_articles(cursor: str = "") -> dict[str, Any]:
        """List articles owned by the authenticated user."""
        return service.list_my_articles(cursor)

    @mcp.tool()
    def bbs_list_comments(
        article_id: int | str | None = None,
        article_url: str | None = None,
        cursor: str = "",
    ) -> dict[str, Any]:
        """List comments and replies belonging to one article."""
        return service.list_comments(article_id, article_url, cursor)

    @mcp.tool()
    def bbs_publish_article(
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
        """Publish an article. Set confirm=true to perform the write."""
        return service.publish_article(
            title,
            content,
            category_id,
            category_path,
            tags,
            summary,
            content_type,
            allow_comments,
            confirm,
        )

    @mcp.tool()
    def bbs_update_article(
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
        """Update an article owned by the authenticated user. Set confirm=true to write."""
        return service.update_article(
            article_id,
            title,
            content,
            category_id,
            category_path,
            tags,
            summary,
            allow_comments,
            confirm,
        )

    @mcp.tool()
    def bbs_set_comment_policy(
        article_id: int | str,
        allow_comments: bool,
        confirm: bool = False,
    ) -> dict[str, Any]:
        """Enable or disable new comments and replies for an article."""
        return service.set_comment_policy(article_id, allow_comments, confirm)

    @mcp.tool()
    def bbs_publish_comment(
        content: str,
        article_id: int | str | None = None,
        article_url: str | None = None,
        reply_to_comment_id: int | str | None = None,
        confirm: bool = False,
    ) -> dict[str, Any]:
        """Publish an article comment or a reply to a comment in that article."""
        return service.publish_comment(
            content,
            article_id,
            article_url,
            reply_to_comment_id,
            confirm,
        )

    @mcp.tool()
    def bbs_delete_comment(comment_id: int | str, confirm: bool = False) -> dict[str, Any]:
        """Delete a comment owned by the authenticated user."""
        return service.delete_comment(comment_id, confirm)

    @mcp.tool()
    def bbs_logout(confirm: bool = False) -> dict[str, Any]:
        """Revoke the bbs-go session and remove the local credential."""
        return service.logout(confirm)

    return mcp


def run_stdio(config: Config | None = None) -> None:
    create_server(config).run(transport="stdio")
