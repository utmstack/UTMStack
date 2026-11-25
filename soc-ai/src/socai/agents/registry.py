"""The specialised agents. Each owns one area of UTMStack, with the tools of that area and a short,
focused prompt; the supervisor decides which one gets a task."""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any

from langchain_core.tools import BaseTool

from ..backend import Backend
from ..config import Settings
from ..generic import mcp_tools
from ..tools import Toolbox, action_tools, read_tools

COMMON = """You work inside UTMStack, a SIEM. You are one specialist among several; a supervisor gave you ONE task.

How to work:
- Do the task and report. Answer with what you found, not with what you did.
- Use the fewest tool calls that settle it. Never repeat a call; its result is already in front of you.
- Never invent data. If a tool fails or comes back empty, say that, and what you would try.
- Tools marked ACTION change things. Use one only when the task asks for it and the evidence is clear.
- Reply in a few short paragraphs or a short list. No preamble."""


@dataclass
class SubAgent:
    name: str
    description: str
    prompt: str
    tools: dict[str, BaseTool] = field(default_factory=dict)


def build_subagents(backend: Backend, settings: Settings, specs: list[dict[str, Any]], *, groups: frozenset[str],
                    tenant: str | None = None, hide_people: bool = False) -> dict[str, SubAgent]:
    """The specialists for a request that asked the backend for these groups. A specialist whose area is
    not on has no tools and is not created, so the supervisor never hears of it."""
    tb = Toolbox(backend, tenant, groups, hide_people)
    reads, acts = read_tools(tb), action_tools(tb)

    def generic(prefixes: tuple[str, ...], group: str) -> dict[str, BaseTool]:
        # The backend already lists only what the groups allow; this refuses an area whose group is off
        # even if a listing said otherwise. "none" is for tools that belong to no group (the core).
        if group != "none" and group not in groups:
            return {}

        return mcp_tools(backend, specs, prefixes, tenant=tenant, allow_changes=True, hide=hide_people)

    def pick(src: dict[str, BaseTool], *names: str) -> dict[str, BaseTool]:
        return {n: src[n] for n in names if n in src}

    alerts_tools = {**pick(reads, "get_alert_dossier", "find_alerts", "top_values", "search_logs"),
                    **pick(acts, "add_alert_note")}

    if "alert_actions" in groups:
        alerts_tools.update(pick(acts, "set_alert_status", "tag_alerts"))

    if "incident_actions" in groups:
        alerts_tools.update(pick(acts, "create_incident"))

    incident_tools = {**pick(reads, "get_incident", "list_incidents", "find_alerts"),
                      **{n: t for n, t in generic(("incident_notes.",), "none").items()
                         if "incident_actions" in groups or not t.description.startswith("ACTION")}}

    if "incident_actions" in groups:
        incident_tools.update(pick(acts, "create_incident"))

    agents = [
        SubAgent("alerts", "Investigate, triage and annotate alerts: what an alert is, what surrounds it, whether it "
                 "looks real, and which other alerts relate to it.",
                 COMMON + "\n\nYou are the alerts specialist. Start from get_alert_dossier when you have an alert id; it "
                 "already holds the score, the related alerts and the recurrence. Reach for find_alerts, top_values or "
                 "search_logs only to resolve a specific doubt.", alerts_tools),
        SubAgent("incidents", "Look at incidents, what is in them and their notes, or open one from alerts.",
                 COMMON + "\n\nYou are the incidents specialist.", incident_tools),
        SubAgent("logs", "Search and aggregate raw log events: who did what from where, counts and top values.",
                 COMMON + "\n\nYou are the log analysis specialist. Field names are exact paths; check them with "
                 "the dataset fields tool before filtering on one you are not sure of.",
                 {**pick(reads, "search_logs", "top_values"), **generic(("store.dataset",), "none")}),
        SubAgent("dashboards", "Create and manage dashboards and visualisations.",
                 COMMON + "\n\nYou are the dashboards specialist. Check every visualisation query returns data "
                 "before you create it, and say plainly which ones you could not make.",
                 {**generic(("dashboards.", "visualizations."), "dashboards"),
                  **(generic(("store.dataset",), "none") if "dashboards" in groups else {})}),
        SubAgent("compliance", "Compliance frameworks, controls and their evaluation.",
                 COMMON + "\n\nYou are the compliance specialist.", generic(("compliance.",), "compliance")),
        SubAgent("soar", "SOAR response rules: list, create, enable and see what ran.",
                 COMMON + "\n\nYou are the SOAR specialist. Build a rule from what you were told, exactly; do not model "
                 "it on existing ones unless asked.", generic(("soar.",), "soar")),
        SubAgent("datasources", "Data sources: what is connected and how much it sends.",
                 COMMON + "\n\nYou are the data sources specialist.",
                 generic(("datasources.", "ingestion_stats."), "datasources")),
        SubAgent("detection", "Correlation rules: see them, create them, switch them on or off.",
                 COMMON + "\n\nYou are the detection engineering specialist.",
                 generic(("correlation_rule.",), "detection")),
    ]

    return {a.name: a for a in agents if a.tools}
