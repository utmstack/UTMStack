import httpx
import pytest

from socai.llm import build_model, headers_for, http_client
from socai.model_config import ModelConfig


def cfg(**kw):
    base = dict(provider="openai", model="m", url="https://llm.example/v1/chat/completions", api_key="sk-1")

    return ModelConfig(**{**base, **kw})


async def sent_headers(c: ModelConfig) -> httpx.Headers:
    seen = {}

    def handler(request: httpx.Request) -> httpx.Response:
        seen["h"] = request.headers

        return httpx.Response(200, json={})

    async with http_client(c, transport=httpx.MockTransport(handler)) as client:
        await client.post("https://llm.example/x", headers={"Authorization": "Bearer sk-1", **headers_for(c)})

    return seen["h"]


async def test_a_bearer_key_is_sent_as_one():
    assert (await sent_headers(cfg()))["authorization"] == "Bearer sk-1"


async def test_with_no_authentication_no_authorization_header_goes_out_and_the_custom_ones_do():
    h = await sent_headers(cfg(auth_type="none", api_key="", headers={"id": "inst", "key": "secret"}))

    assert "authorization" not in h and h["id"] == "inst" and h["key"] == "secret"


async def test_a_key_that_goes_in_a_header_of_its_own_is_not_also_sent_as_a_bearer_token():
    h = await sent_headers(cfg(auth_type="header", auth_header_name="x-api-key"))

    assert "authorization" not in h and h["x-api-key"] == "sk-1"


def test_the_endpoint_the_settings_page_holds_becomes_the_base_the_library_wants():
    m = build_model(cfg())

    assert str(m.openai_api_base) == "https://llm.example/v1"


def test_a_provider_with_other_headers_and_no_key_builds():
    m = build_model(cfg(provider="threatwinds", url="https://cm/proxy/api/ai/v1/chat/completions", api_key="",
                        auth_type="none", headers={"id": "i", "key": "k"}))

    assert str(m.openai_api_base) == "https://cm/proxy/api/ai/v1"


def test_anthropic_builds_from_its_messages_endpoint():
    m = build_model(cfg(provider="anthropic", url="https://api.anthropic.com/v1/messages"))

    assert "anthropic.com" in str(m.anthropic_api_url) and not str(m.anthropic_api_url).endswith("/messages")


def test_an_incomplete_configuration_is_refused_before_it_is_used():
    with pytest.raises(ValueError):
        build_model(ModelConfig(provider="openai", model="m"))
