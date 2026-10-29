from bbs_go_mcp.config import Config
from bbs_go_mcp.credential_store import Credential, CredentialStore
from bbs_go_mcp.tool_service import ToolService


class FakeClient:
    def __init__(self, config, credential):
        self.config = config
        self.credential = credential

    def __enter__(self):
        return self

    def __exit__(self, *_):
        return None

    def list_categories(self, request_id=""):
        return [{"id": 2, "name": "技术", "path": "技术", "allowPost": True}]

    def get_article_edit(self, article_id, request_id=""):
        return {
            "id": article_id,
            "title": "title",
            "categoryId": 2,
            "content": "content",
            "tags": ["go"],
            "cover": None,
            "allowComments": True,
        }

    def update_article(self, article_id, body, request_id=""):
        return {"articleId": article_id, "body": body}


class FakeStore(CredentialStore):
    def load(self):
        return Credential("cookie")


def test_update_article_confirm_false_does_not_write():
    service = ToolService(
        Config(),
        FakeStore(use_keyring=False),
        client_factory=FakeClient,
    )
    result = service.update_article(1, allow_comments=False, confirm=False)
    assert result["error"] == "CONFIRMATION_REQUIRED"
    assert result["requiredConfirmation"] is True


def test_update_article_confirm_true_writes():
    service = ToolService(
        Config(),
        FakeStore(use_keyring=False),
        client_factory=FakeClient,
    )
    result = service.update_article(1, allow_comments=False, confirm=True)
    assert result["success"] is True
    assert result["data"]["body"]["allowComments"] is False
