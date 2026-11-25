"""One agent: a model that calls tools until it can answer, with limits that stop it wandering.

Agents that "make a lot of calls and never get anywhere" share three habits: they repeat a call whose
answer they already have, they keep calling after they could have answered, and they never run out of
road. This loop breaks each one: a repeated call is answered with a pointer to its first result, the
total of calls and steps is capped, and when the cap is near the model is told to answer with what it
has.
"""

from __future__ import annotations

import asyncio
import json
from typing import Any

from langchain_core.language_models.chat_models import BaseChatModel
from langchain_core.messages import AIMessage, SystemMessage, ToolMessage
from langchain_core.tools import BaseTool
from langgraph.graph import END, START, MessagesState, StateGraph


class AgentState(MessagesState):
    steps: int
    calls: int
    seen: dict[str, bool]


REPEAT = ("You already made this exact call and its result is earlier in the conversation. Use it, or ask "
          "for something different.")
OUT_OF_CALLS = "No more tool calls are available. Answer now with what you have, and say what you could not check."
WRAP_UP = ("You have no steps left. Answer now with the information you already have. If something is missing, say "
           "so plainly; do not guess.")


def _key(call: dict[str, Any]) -> str:
    return call["name"] + json.dumps(call.get("args") or {}, sort_keys=True, default=str)


def react(model: BaseChatModel, tools: dict[str, BaseTool], system: str, *, max_steps: int = 10,
          max_tool_calls: int = 14):
    bound = model.bind_tools(list(tools.values())) if tools else model

    async def agent(state: AgentState) -> dict[str, Any]:
        steps = state.get("steps", 0)
        messages = [SystemMessage(system), *state["messages"]]

        if steps >= max_steps - 1:
            reply = await model.ainvoke([*messages, SystemMessage(WRAP_UP)])
        else:
            reply = await bound.ainvoke(messages)

        return {"messages": [reply], "steps": steps + 1}

    async def act(state: AgentState) -> dict[str, Any]:
        last = state["messages"][-1]
        calls = state.get("calls", 0)
        seen = dict(state.get("seen") or {})
        pending: list[tuple[dict[str, Any], asyncio.Future | None, str | None]] = []

        for call in last.tool_calls:
            key = _key(call)

            if key in seen:
                pending.append((call, None, REPEAT))
            elif calls >= max_tool_calls:
                pending.append((call, None, OUT_OF_CALLS))
            elif call["name"] not in tools:
                pending.append((call, None, f"ERROR: there is no tool named {call['name']}"))
            else:
                seen[key] = True
                calls += 1
                pending.append((call, asyncio.ensure_future(tools[call["name"]].ainvoke(call.get("args") or {})), None))

        out: list[ToolMessage] = []

        for call, fut, canned in pending:
            if fut is None:
                content = canned or ""
            else:
                try:
                    content = str(await fut)
                except Exception as exc:  # noqa: BLE001 - a failing tool is information for the model
                    content = f"ERROR: {exc}"

            out.append(ToolMessage(content=content, tool_call_id=call["id"], name=call["name"]))

        return {"messages": out, "calls": calls, "seen": seen}

    def after_agent(state: AgentState) -> str:
        last = state["messages"][-1]

        return "act" if isinstance(last, AIMessage) and last.tool_calls else END

    g = StateGraph(AgentState)
    g.add_node("agent", agent)
    g.add_node("act", act)
    g.add_edge(START, "agent")
    g.add_conditional_edges("agent", after_agent, {"act": "act", END: END})
    g.add_edge("act", "agent")

    return g.compile()


def last_text(state: dict[str, Any]) -> str:
    """What the agent finally said."""
    for m in reversed(state["messages"]):
        if isinstance(m, AIMessage) and not m.tool_calls:
            c = m.content

            return c if isinstance(c, str) else "".join(p.get("text", "") for p in c if isinstance(p, dict))

    return ""
