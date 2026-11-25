"""What the API and the worker share: the settings, the backend, the model and the tools published."""

from __future__ import annotations

import json
import time
from dataclasses import asdict, dataclass, field, replace
from typing import Any

from langchain_core.language_models.chat_models import BaseChatModel

from .agents.registry import SubAgent, build_subagents
from .backend import Backend
from .config import Settings
from .config_source import ConfigSource
from .llm import build_model
from .supervisor import supervisor
from .triage import triage_graph

SPECS_TTL = 300.0


class NotConfigured(Exception):
    """No model is configured for the tenant, or the backend has not published the configuration yet."""


@dataclass
class Runtime:
    settings: Settings
    backend: Backend
    # A fixed model, for development and tests; otherwise the model comes from the tenant's configuration.
    model: BaseChatModel | None = None
    configs: ConfigSource | None = None
    _specs: dict[frozenset[str], tuple[float, list[dict[str, Any]]]] = field(default_factory=dict)
    _models: dict[str, BaseChatModel] = field(default_factory=dict)

    def model_for(self, tenant: str | None) -> tuple[BaseChatModel, Settings]:
        """The model of a tenant and the limits its configuration sets."""
        if self.model is not None:
            return self.model, self.settings

        cfg = self.configs.get(tenant) if self.configs else None

        if cfg is None:
            raise NotConfigured("SOC AI is not configured for this tenant")

        key = json.dumps(asdict(cfg), sort_keys=True)

        if key not in self._models:
            self._models[key] = build_model(cfg, temperature=self.settings.llm_temperature)

        triage = set(self.settings.triage_groups)

        if cfg.allow_incidents:
            triage.add("incident_actions")

        return self._models[key], replace(self.settings, max_steps=cfg.max_steps, triage_groups=frozenset(triage))

    async def specs(self, groups: frozenset[str]) -> list[dict[str, Any]]:
        """The tools the backend shows for these groups, remembered for a few minutes."""
        hit = self._specs.get(groups)

        if hit is None or time.monotonic() - hit[0] > SPECS_TTL:
            hit = (time.monotonic(), await self.backend.list_tools(groups))
            self._specs[groups] = hit

        return hit[1]

    async def agents(self, *, tenant: str | None, groups: frozenset[str], hide_people: bool = False,
                     settings: Settings | None = None) -> dict[str, SubAgent]:
        return build_subagents(self.backend, settings or self.settings, await self.specs(groups), groups=groups,
                               tenant=tenant, hide_people=hide_people)

    async def chat_graph(self, *, tenant: str | None, groups: frozenset[str] | None = None):
        model, settings = self.model_for(tenant)
        on = settings.chat_groups if groups is None else groups
        agents = await self.agents(tenant=tenant, groups=on, settings=settings)

        return supervisor(model, agents, settings)

    async def triage(self, alert_id: str, tenant: str | None, facts: dict[str, Any] | None = None,
                     force: bool = False) -> str:
        model, settings = self.model_for(tenant)
        agents = await self.agents(tenant=tenant, groups=settings.triage_groups, hide_people=True, settings=settings)
        graph = triage_graph(model, self.backend, settings, agents, tenant=tenant)
        state: dict[str, Any] = {"alert_id": alert_id}

        if facts:
            state["facts"] = facts

        if force:
            state["force"] = True

        out = await graph.ainvoke(state)

        return out.get("outcome", "failed: no outcome")
