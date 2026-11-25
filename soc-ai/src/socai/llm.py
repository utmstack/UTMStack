"""The chat model for one configuration, whichever provider it names.

A configuration says where the model is (a full endpoint, as the settings page asks for it), how to
authenticate (a bearer key, a key in a header of its own, or nothing) and which headers to add. The
instance default is ThreatWinds' proxy, which authenticates with two headers and no key.
"""

from __future__ import annotations

import httpx
from langchain_core.language_models.chat_models import BaseChatModel

from .model_config import ModelConfig

TIMEOUT = 120.0


def _base(url: str, suffix: str) -> str:
    """The base address a client library wants, from the full endpoint the configuration holds."""
    u = url.rstrip("/")

    return u[: -len(suffix)] if u.endswith(suffix) else u


async def _strip_authorization(request: httpx.Request) -> None:
    request.headers.pop("Authorization", None)


def http_client(cfg: ModelConfig, *, transport: httpx.AsyncBaseTransport | None = None) -> httpx.AsyncClient:
    """The HTTP client for an OpenAI-compatible endpoint. The library always sends the key as a bearer
    token; when the configuration says the key goes elsewhere, or there is none, that header is removed."""
    hooks = [] if cfg.auth_type == "bearer" else [_strip_authorization]

    return httpx.AsyncClient(timeout=TIMEOUT, transport=transport, event_hooks={"request": hooks})


def headers_for(cfg: ModelConfig) -> dict[str, str]:
    out = dict(cfg.headers)

    if cfg.auth_type == "header" and cfg.auth_header_name and cfg.api_key:
        out[cfg.auth_header_name] = cfg.api_key

    return out


def build_model(cfg: ModelConfig, *, temperature: float = 0.0) -> BaseChatModel:
    if not cfg.active:
        raise ValueError("the model configuration is incomplete: provider, model and url are needed")

    if cfg.provider == "anthropic":
        from langchain_anthropic import ChatAnthropic

        return ChatAnthropic(model=cfg.model, api_key=cfg.api_key or "none", base_url=_base(cfg.url, "/v1/messages"),
                             default_headers=headers_for(cfg) or None, temperature=temperature,
                             max_tokens=cfg.max_tokens, timeout=TIMEOUT)

    from langchain_openai import ChatOpenAI

    return ChatOpenAI(model=cfg.model, api_key=cfg.api_key if cfg.auth_type == "bearer" and cfg.api_key else "unused",
                      base_url=_base(cfg.url, "/chat/completions"), default_headers=headers_for(cfg) or None,
                      http_async_client=http_client(cfg), temperature=temperature, max_tokens=cfg.max_tokens,
                      timeout=TIMEOUT)
