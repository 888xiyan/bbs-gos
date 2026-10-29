import json

import pytest

from bbs_go_mcp.config import Config, load_config, save_config
from bbs_go_mcp.errors import BbsMcpError


def test_config_round_trip(tmp_path):
    path = tmp_path / "config.json"
    config = Config(
        baseUrl="https://bbs.example.test",
        timeoutSeconds=45,
        defaultConfirm=True,
        allowedArticleUrlHosts=["BBS.EXAMPLE.TEST"],
        maxContentLength=1000,
    )
    save_config(config, path)
    loaded = load_config(path)
    assert loaded.baseUrl == config.baseUrl
    assert loaded.timeoutSeconds == 45
    assert loaded.defaultConfirm is True
    assert loaded.allowedArticleUrlHosts == ["bbs.example.test"]
    assert json.loads(path.read_text())["maxContentLength"] == 1000


@pytest.mark.parametrize("url", ["", "ftp://example.test", "example.test"])
def test_invalid_base_url_rejected(url, tmp_path):
    with pytest.raises(BbsMcpError):
        save_config(Config(baseUrl=url), tmp_path / "config.json")
