"""The model a tenant's alerts and chats run on.

The backend keeps one configuration for the instance and, for each tenant that has set its own, another;
a tenant without one uses the instance's. It publishes all of them in one document (see
backend/modules/socai/config_mirror.go); this reads it.
"""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from typing import Any

from .crypto import Cipher

PROVIDER_URLS = {
    "openai": "https://api.openai.com/v1/chat/completions",
    "anthropic": "https://api.anthropic.com/v1/messages",
    "gemini": "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
    "ollama": "http://localhost:11434/v1/chat/completions",
    "mistral": "https://api.mistral.ai/v1/chat/completions",
    "deepseek": "https://api.deepseek.com/chat/completions",
    "groq": "https://api.groq.com/openai/v1/chat/completions",
}

DEFAULT_MAX_TOKENS = 4096
DEFAULT_MAX_STEPS = 12


@dataclass(frozen=True)
class ModelConfig:
    provider: str = ""
    model: str = ""
    url: str = ""
    api_key: str = ""
    auth_type: str = "bearer"  # bearer | header | none
    auth_header_name: str = ""
    headers: dict[str, str] = field(default_factory=dict)
    max_tokens: int = DEFAULT_MAX_TOKENS
    max_steps: int = DEFAULT_MAX_STEPS
    auto_analyze: bool = False
    allow_incidents: bool = False

    @property
    def active(self) -> bool:
        return bool(self.provider and self.model and self.url)


def _one(raw: dict[str, Any], cipher: Cipher | None) -> ModelConfig:
    def dec(v: str) -> str:
        return cipher.decrypt(v) if cipher and v else v

    provider = (raw.get("provider") or "").lower()
    url = raw.get("url") or PROVIDER_URLS.get(provider, "")

    return ModelConfig(
        provider=provider,
        model=raw.get("model") or "",
        url=url,
        api_key=dec(raw.get("api_key") or ""),
        auth_type=raw.get("auth_type") or "bearer",
        auth_header_name=raw.get("auth_header_name") or "",
        headers={k: dec(v) for k, v in (raw.get("custom_headers") or {}).items()},
        max_tokens=int(raw.get("max_tokens") or 0) or DEFAULT_MAX_TOKENS,
        max_steps=int(raw.get("max_tool_iterations") or 0) or DEFAULT_MAX_STEPS,
        auto_analyze=bool(raw.get("auto_analyze")),
        allow_incidents=bool(raw.get("allow_incidents")),
    )


@dataclass(frozen=True)
class ModelConfigs:
    default: ModelConfig | None = None
    tenants: dict[str, ModelConfig] = field(default_factory=dict)

    def for_tenant(self, tenant: str | None) -> ModelConfig | None:
        """The tenant's own, else the instance's. None when nothing usable is configured."""
        cfg = self.tenants.get(tenant) if tenant else None

        if cfg is None:
            cfg = self.default

        return cfg if cfg is not None and cfg.active else None

    @classmethod
    def parse(cls, document: str, cipher: Cipher | None) -> "ModelConfigs":
        raw = json.loads(document)

        return cls(
            default=_one(raw["default"], cipher) if raw.get("default") else None,
            tenants={t: _one(c, cipher) for t, c in (raw.get("tenants") or {}).items()},
        )
