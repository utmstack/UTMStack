"""Settings, read from the environment."""

from __future__ import annotations

import os
from dataclasses import dataclass, field


GROUPS = frozenset({"alert_actions", "incident_actions", "dashboards", "compliance", "soar", "detection",
                    "datasources"})

# The names the settings page and the old plugin used.
LEGACY = {"alerts": "alert_actions", "incidents": "incident_actions", "correlation": "detection"}


def normalize_groups(names) -> frozenset[str]:
    out = {LEGACY.get(n.strip(), n.strip()) for n in names if n and n.strip()}

    return frozenset(g for g in out if g in GROUPS)


def _csv(name: str, default: str = "") -> frozenset[str]:
    return normalize_groups(os.getenv(name, default).split(","))


@dataclass(frozen=True)
class Settings:
    backend_url: str = "http://backend:8080"
    internal_key: str = ""
    nats_url: str = ""
    http_port: int = 8090

    # The model. "openai" also covers any OpenAI-compatible endpoint through base_url.
    llm_provider: str = "openai"
    llm_model: str = ""
    llm_api_key: str = ""
    llm_base_url: str = ""
    llm_temperature: float = 0.0
    llm_max_tokens: int = 2048

    # What decrypts the model configuration the backend publishes in NATS: the backend's own key.
    encryption_key: str = ""

    # The capability groups each kind of work asks the backend for. The backend shows the agent only its
    # core (what investigating an alert takes, and the assessment note) plus these, so what is not asked
    # for does not exist for it. Groups: alert_actions, incident_actions, dashboards, compliance, soar,
    # detection, datasources.
    #
    # The automatic triage asks for none by default; an administrator allows it to open incidents
    # (incident_actions) or to close and tag alerts (alert_actions). The chat has a person driving each
    # conversation, so it gets every group.
    triage_groups: frozenset[str] = field(default_factory=frozenset)
    chat_groups: frozenset[str] = field(default_factory=lambda: GROUPS)

    max_steps: int = 10
    max_tool_calls: int = 14
    triage_concurrency: int = 4

    @classmethod
    def from_env(cls) -> "Settings":
        return cls(
            backend_url=os.getenv("BACKEND_URL", "http://backend:8080").rstrip("/"),
            internal_key=os.getenv("INTERNAL_KEY", ""),
            nats_url=os.getenv("NATS_URL", ""),
            http_port=int(os.getenv("HTTP_PORT", "8090")),
            llm_provider=os.getenv("LLM_PROVIDER", "openai").lower(),
            llm_model=os.getenv("LLM_MODEL", ""),
            llm_api_key=os.getenv("LLM_API_KEY", ""),
            llm_base_url=os.getenv("LLM_BASE_URL", ""),
            llm_temperature=float(os.getenv("LLM_TEMPERATURE", "0")),
            llm_max_tokens=int(os.getenv("LLM_MAX_TOKENS", "2048")),
            encryption_key=os.getenv("ENCRYPTION_KEY", ""),
            triage_groups=_csv("TRIAGE_GROUPS"),
            chat_groups=_csv("CHAT_GROUPS", ",".join(sorted(GROUPS))),
            max_steps=int(os.getenv("MAX_STEPS", "10")),
            max_tool_calls=int(os.getenv("MAX_TOOL_CALLS", "14")),
            triage_concurrency=int(os.getenv("TRIAGE_CONCURRENCY", "4")),
        )
