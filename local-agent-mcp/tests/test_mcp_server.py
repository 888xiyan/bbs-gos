from bbs_go_mcp.server import create_server


def test_server_registers_expected_tools():
    server = create_server()
    names = set(server._tool_manager._tools.keys())
    expected = {
        "bbs_status",
        "bbs_auth_status",
        "bbs_list_categories",
        "bbs_get_article",
        "bbs_list_my_articles",
        "bbs_list_comments",
        "bbs_publish_article",
        "bbs_update_article",
        "bbs_set_comment_policy",
        "bbs_publish_comment",
        "bbs_delete_comment",
        "bbs_logout",
    }
    assert expected.issubset(names)
