"""The master agent: reads what the user asked, hands narrow tasks to the specialists, and answers.

It does not call tools of its own. Its only decisions are "ask this specialist this" and "I can answer
now", which is what keeps a question from turning into a long wander over a hundred tools.
"""

from __future__ import annotations

import re
from typing import Any, Literal

from langchain_core.language_models.chat_models import BaseChatModel
from langchain_core.messages import AIMessage, HumanMessage, SystemMessage
from langgraph.graph import END, START, StateGraph
from pydantic import BaseModel, Field
from typing_extensions import TypedDict

from .agents.registry import SubAgent
from .decide import decide
from .graph import last_text, react
from .config import Settings


class Route(BaseModel):
    """The next step."""

    action: Literal["delegate", "finish"] = Field(description="delegate to a specialist, or finish with the answer")
    agent: str | None = Field(None, description="The specialist, when delegating")
    task: str | None = Field(None, description="What to ask it: self-contained, with every id and name it needs")
    answer: str | None = Field(None, description="The reply to the user, when finishing")


class SupState(TypedDict, total=False):
    messages: list[Any]
    context: str  # where the user is in the UI, if they said
    lang: str | None
    findings: list[str]
    hops: int
    asked: list[tuple[str, str]]  # (specialist, task) already handed out
    answer: str
    pending: Route | None


def _system(agents: dict[str, SubAgent], context: str, lang: str | None) -> str:
    roster = "\n".join(f"- {a.name}: {a.description}" for a in agents.values())
    language = (f'Write the answer in the language with ISO code "{lang}".' if lang
                else "Answer in the language the user wrote in.")

    return f"""You are the UTMStack operations agent, the supervisor of a team of specialists. The user talks to you; \
you get their request done by asking the specialists and then answering.

Specialists:
{roster}

{context}

Rules:
- If you can answer without looking anything up, finish at once with the answer.
- Otherwise delegate ONE narrow, self-contained task at a time, with every id and name it needs. Read what comes back \
before deciding the next step.
- Never delegate the same thing twice. When the findings are enough, finish.
- A specialist may say a change is not enabled. Report that; do not insist.
- If a specialist could not do something, do not ask again in other words: say what is missing.
- Never mention specialists, delegations, budgets, steps or tools to the user. Say what you could not find or do \
in terms of the data and the request.
- Finish with a direct answer: what you found or did, and what you could not do. Never invent data.
{language}"""


_WORD = re.compile(r"\w+", re.UNICODE)


def _already_asked(asked: list[tuple[str, str]], agent: str, task: str) -> bool:
    """The same specialist given a task that says much the same: most of its words were in an earlier one."""
    words = {w.lower() for w in _WORD.findall(task) if len(w) > 2}

    if not words:
        return False

    for who, earlier in asked:
        if who != agent:
            continue

        before = {w.lower() for w in _WORD.findall(earlier) if len(w) > 2}

        if before and len(words & before) / len(words | before) >= 0.6:
            return True

    return False


def supervisor(model: BaseChatModel, agents: dict[str, SubAgent], settings: Settings, *, max_hops: int = 4):
    workers = {
        n: react(model, a.tools, a.prompt, max_steps=settings.max_steps, max_tool_calls=settings.max_tool_calls)
        for n, a in agents.items()
    }

    async def route(state: SupState) -> dict[str, Any]:
        hops = state.get("hops", 0)
        findings = state.get("findings", [])
        system = _system(agents, state.get("context", ""), state.get("lang"))
        notes = "\n\n".join(findings) if findings else "(nothing yet)"
        messages = [SystemMessage(system), *state["messages"], SystemMessage(f"Findings so far:\n{notes}")]

        if hops >= max_hops:
            messages.append(SystemMessage("You have used all your delegations. Finish now."))

        r = await decide(model, Route, messages)
        asked = state.get("asked", [])

        if r.action == "delegate" and r.agent in agents and r.task and _already_asked(asked, r.agent, r.task):
            # The same question to the same specialist in other words gets the same answer: the reply is
            # already in the findings. Say so, and spend a delegation so the loop has an end.
            note = (f"[supervisor] you already asked {r.agent} this; its answer is in the findings above. Finish with what "
                    "you have and say what could not be found.")
            return {"findings": [*findings, note], "hops": hops + 1, "pending": None}

        if r.action == "delegate" and (hops >= max_hops or r.agent not in agents or not r.task):
            note = (f"[supervisor] there is no specialist named {r.agent!r}" if r.agent not in agents
                    else "[supervisor] delegation refused")
            return {"findings": [*findings, note], "hops": hops + 1, "pending": None}

        return {"pending": r, "asked": [*asked, (r.agent, r.task)] if r.action == "delegate" and r.agent and r.task else asked}

    def after_route(state: SupState) -> str:
        r = state.get("pending")

        if r is None:
            return "route"

        return "delegate" if r.action == "delegate" else "finish"

    async def delegate(state: SupState) -> dict[str, Any]:
        r = state["pending"]
        assert r is not None and r.agent is not None and r.task is not None
        out = await workers[r.agent].ainvoke({"messages": [HumanMessage(r.task)]})
        text = last_text(out) or "(it returned nothing)"

        return {"findings": [*state.get("findings", []), f"[{r.agent}] {text}"], "hops": state.get("hops", 0) + 1,
                "pending": None}

    async def finish(state: SupState) -> dict[str, Any]:
        r = state["pending"]
        answer = (r.answer if r else "") or ""

        if not answer.strip():
            findings = "\n\n".join(state.get("findings", [])) or "(nothing was found)"
            reply = await model.ainvoke([
                SystemMessage("Write the final answer to the user from these findings. Do not invent anything."),
                *state["messages"], SystemMessage(findings)])
            answer = reply.content if isinstance(reply.content, str) else ""

        return {"answer": answer, "messages": [*state["messages"], AIMessage(answer)]}

    g = StateGraph(SupState)
    g.add_node("route", route)
    g.add_node("delegate", delegate)
    g.add_node("finish", finish)
    g.add_edge(START, "route")
    g.add_conditional_edges("route", after_route, {"route": "route", "delegate": "delegate", "finish": "finish"})
    g.add_edge("delegate", "route")
    g.add_edge("finish", END)

    return g.compile()
