import pytest

from bbs_go_mcp.category_resolver import normalize_category_path, resolve_category
from bbs_go_mcp.errors import BbsMcpError, ErrorCode

CATEGORIES = [
    {
        "id": 1,
        "name": "技术",
        "path": "技术",
        "allowPost": True,
        "children": [
            {
                "id": 2,
                "name": "后端",
                "path": "技术 / 后端",
                "allowPost": True,
                "children": [{"id": 3, "name": "Go", "path": "技术 / 后端 / Go", "allowPost": True}],
            }
        ],
    }
]


def test_resolve_category_by_id_and_path():
    assert resolve_category(CATEGORIES, category_id=3)["id"] == 3
    assert resolve_category(CATEGORIES, category_path="技术 / 后端 / Go")["id"] == 3
    assert normalize_category_path(" 技术 /后端/ Go ") == "技术 / 后端 / Go"


def test_category_errors():
    with pytest.raises(BbsMcpError) as not_found:
        resolve_category(CATEGORIES, category_path="不存在")
    assert not_found.value.code == ErrorCode.CATEGORY_NOT_FOUND

    disabled = [{"id": 9, "name": "关闭", "path": "关闭", "allowPost": False}]
    with pytest.raises(BbsMcpError) as not_postable:
        resolve_category(disabled, category_id=9)
    assert not_postable.value.code == ErrorCode.CATEGORY_NOT_POSTABLE
