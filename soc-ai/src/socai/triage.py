"""Triage of one alert, from the moment it exists to a note on it.

A pipeline, not a free agent: the facts are fetched in one call, a model gives a verdict in one call, and
only if it says it cannot tell does a specialist investigate, once, on the precise doubt. Then the verdict
is written as the note the alerts page parses. A free agent given the same alert wanders; this one has
four steps.
"""

from __future__ import annotations

from typing import Any, Literal

from langchain_core.language_models.chat_models import BaseChatModel
from langchain_core.messages import HumanMessage, SystemMessage
from langgraph.graph import END, START, StateGraph
from pydantic import BaseModel, Field
from typing_extensions import TypedDict

from .agents.registry import SubAgent
from .anonymize import hide_people
from .backend import Backend, ToolError
from .config import Settings
from .decide import decide
from .graph import last_text, react
from .shaping import render
from .tools import Toolbox

MARK = "[AI SOC Agent]"


class Verdict(BaseModel):
    """The assessment of one alert."""

    classification: Literal["possible incident", "possible false positive", "standard alert"]
    confident: bool = Field(description="False only when the alert and its score are not enough to decide and a "
                            "specific lookup would settle it. Do not use it out of caution.")
    missing: str = Field("", description="When not confident: exactly what is missing and where it would be found")
    adjustment: int = Field(0, ge=-10, le=10, description="Points to move the given score by: 0 unless the facts clearly say "
                            "otherwise, never more than 10 either way. Say why in context.")
    score: int | None = Field(None, ge=0, le=100, description="Only when no score was given with the facts: your own, 0 to 100")
    risk: Literal["Low", "Medium", "High"] | None = Field(None, description="Ignored when a score was given: it follows the score")
    threat_assessment: str = Field(description="One sentence")
    affected_asset: str = "N/A"
    context: str = Field(description="What the score breakdown and the alert show")
    analysis: str = Field(description="Your reasoning, short")
    action: str = Field(description="The recommended next step")


def _clean(text: str) -> str:
    return " ".join(str(text).replace("|", "/").split())


def final_score(v: Verdict, facts: dict[str, Any]) -> int:
    """The score on the note: the backend's, moved by at most ten points.

    The model used to write its own, and the same alert came out anywhere between 22 and 33 while the context
    still quoted 33. The base is worked out by the backend and is the same every time; the model only decides how
    far the facts move it."""
    base = (facts.get("score") or {}).get("score") if isinstance(facts.get("score"), dict) else None

    if isinstance(base, (int, float)):
        return max(0, min(100, round(base) + v.adjustment))

    return v.score if v.score is not None else 50


def risk_for(score: int) -> str:
    return "High" if score >= 70 else "Medium" if score >= 40 else "Low"


def format_note(v: Verdict, score: int | None = None) -> str:
    """The single line the alerts page parses. The order and the separators are its contract."""
    state = "Completed" if v.classification == "possible false positive" else "Open"
    score = score if score is not None else (v.score if v.score is not None else 50)

    return (f"{MARK} Score: {score}/100 - {state} - {risk_for(score)} Risk | Threat Assessment: {_clean(v.threat_assessment)} "
            f"| Affected Asset: {_clean(v.affected_asset)} | Context: {_clean(v.context)} "
            f"| LLM Analysis: {_clean(v.analysis)} | Action: {_clean(v.action)}")


SYSTEM = f"""You are an autonomous SOC analyst inside UTMStack, a SIEM. You are given the facts about ONE alert: its \
fields, its deterministic 0-100 risk score with the breakdown (intrinsic risk, deviation from baseline, asset \
criticality, attack chain), the other alerts around it, and how often it has fired. The score is consistent and \
explainable: it is the starting point. You do not write a new score: you give an adjustment of at most 10 points up \
or down, only when the facts clearly say so, and say why in the context.

Some values are anonymized (names and addresses of people); "anonymizedFields" lists them. Read nothing into \
the placeholders.

Classify it as "possible incident", "possible false positive" or "standard alert". In most cases the facts are \
enough. Say you are not confident only when a specific lookup would change your answer."""


class TriageState(TypedDict, total=False):
    alert_id: str
    facts: dict[str, Any]
    force: bool  # a person asked for this one: assess it again even if it was assessed before
    verdict: Verdict
    investigation: str
    outcome: str  # written | skipped | failed


def triage_graph(model: BaseChatModel, backend: Backend, settings: Settings, agents: dict[str, SubAgent], *,
                 tenant: str | None = None):
    tb = Toolbox(backend, tenant, settings.triage_groups, hide_people=True)
    investigator = agents.get("alerts")
    worker = (react(model, investigator.tools, investigator.prompt, max_steps=settings.max_steps,
                    max_tool_calls=settings.max_tool_calls) if investigator else None)

    async def load(state: TriageState) -> dict[str, Any]:
        # The backend sends the dossier with the alert, worked out next to the data. Only when it could
        # not (or for a manual triage) is it built here, with several calls.
        facts = state.get("facts")

        if not facts:
            try:
                facts = await tb.alert_dossier(state["alert_id"])
            except ToolError as exc:
                return {"outcome": f"failed: {exc}"}

        if not state.get("force") and MARK in str((facts.get("alert") or {}).get("notes") or ""):
            return {"outcome": "skipped: already assessed"}

        facts, hidden = hide_people(facts)

        if hidden:
            facts["anonymizedFields"] = sorted(set(hidden))

        return {"facts": facts}

    def after_load(state: TriageState) -> str:
        # An outcome here means there is nothing to judge: already assessed, or the alert could not be read.
        return END if state.get("outcome") else "verdict"

    async def verdict(state: TriageState) -> dict[str, Any]:
        extra = f"\n\nWhat an investigation found:\n{state['investigation']}" if state.get("investigation") else ""
        final = "\nThis is the final assessment: decide with what you have; do not ask for more." \
            if state.get("investigation") else ""
        v = await decide(model, Verdict,
                         [SystemMessage(SYSTEM + final), HumanMessage(render(state["facts"], max_chars=9000) + extra)])

        if state.get("investigation"):
            v.confident = True

        return {"verdict": v}

    def after_verdict(state: TriageState) -> str:
        v = state["verdict"]

        return "investigate" if (not v.confident and worker is not None and not state.get("investigation")) else "write"

    async def investigate(state: TriageState) -> dict[str, Any]:
        v = state["verdict"]
        task = (f"Alert {state['alert_id']}. A first assessment could not decide because: {v.missing or 'unclear'}. "
                f"Resolve exactly that, with the fewest lookups, and report what you found.")
        out = await worker.ainvoke({"messages": [HumanMessage(task)]})  # type: ignore[union-attr]

        return {"investigation": last_text(out) or "(the investigation returned nothing)"}

    async def write(state: TriageState) -> dict[str, Any]:
        v = state["verdict"]
        alert_id = state["alert_id"]

        try:
            await tb.add_alert_note(alert_id, format_note(v, final_score(v, state["facts"])))

            # Only what an administrator turned on, and only when the verdict is firm.
            if v.confident:
                if v.classification == "possible false positive" and "alert_actions" in settings.triage_groups:
                    await tb.set_alert_status([alert_id], "closed", "Closed by the AI SOC agent: possible false positive")
                elif v.classification == "possible incident" and "incident_actions" in settings.triage_groups:
                    await tb.create_incident([alert_id], f"AI SOC: {(v.threat_assessment or 'possible incident')[:80]}")
        except ToolError as exc:
            return {"outcome": f"failed: {exc}"}

        return {"outcome": "written"}

    g = StateGraph(TriageState)
    g.add_node("load", load)
    g.add_node("verdict", verdict)
    g.add_node("investigate", investigate)
    g.add_node("write", write)
    g.add_edge(START, "load")
    g.add_conditional_edges("load", after_load, {"verdict": "verdict", END: END})
    g.add_conditional_edges("verdict", after_verdict, {"investigate": "investigate", "write": "write"})
    g.add_edge("investigate", "verdict")
    g.add_edge("write", END)

    return g.compile()
