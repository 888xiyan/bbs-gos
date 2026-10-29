import pytest

from bbs_go_mcp.config import Config
from bbs_go_mcp.errors import BbsMcpError
from bbs_go_mcp.security import resolve_article_id, validate_content


def test_article_id_and_allowed_url():
    config = Config(allowedArticleUrlHosts=["bbs.example.test"])
    assert resolve_article_id("12", None, config) == 12
    assert resolve_article_id(None, "https://bbs.example.test/article/34", config) == 34


@pytest.mark.parametrize(
    "url",
    [
        "https://evil.example/article/1",
        "https://bbs.example.test/article/1?x=1",
        "file:///article/1",
    ],
)
def test_article_url_rejected(url):
    config = Config(allowedArticleUrlHosts=["bbs.example.test"])
    with pytest.raises(BbsMcpError):
        resolve_article_id(None, url, config)


def test_content_length_validated():
    with pytest.raises(BbsMcpError):
        validate_content("123456", Config(maxContentLength=5))
