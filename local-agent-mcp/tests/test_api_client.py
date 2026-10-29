import httpx
import pytest
import respx

from bbs_go_mcp.api_client import BbsApiClient
from bbs_go_mcp.config import Config
from bbs_go_mcp.credential_store import Credential
from bbs_go_mcp.errors import BbsMcpError, ErrorCode


@respx.mock
def test_api_client_parses_envelope_and_cookie():
    route = respx.get("https://bbs.example.test/api/user/current").mock(
        return_value=httpx.Response(200, json={"success": True, "data": {"id": "1"}})
    )
    with BbsApiClient(
        Config(baseUrl="https://bbs.example.test"),
        Credential("cookie-value"),
    ) as client:
        assert client.get_current_user("req-1")["id"] == "1"
    assert route.called
    assert route.calls[0].request.headers["cookie"] == "bbsgo_token=cookie-value"


@respx.mock
def test_api_client_maps_auth_errors():
    respx.get("https://bbs.example.test/api/user/current").mock(
        return_value=httpx.Response(401, json={"success": False, "message": "not login"})
    )
    with BbsApiClient(
        Config(baseUrl="https://bbs.example.test"),
        Credential("cookie-value"),
    ) as client, pytest.raises(BbsMcpError) as error:
        client.get_current_user()
    assert error.value.code == ErrorCode.AUTH_REQUIRED
