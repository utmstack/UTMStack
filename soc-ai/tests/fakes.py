"""A model that follows a script and a backend that records what it was asked."""

from __future__ import annotations

from typing import Any, Callable

from langchain_core.language_models.chat_models import BaseChatModel
from langchain_core.messages import AIMessage, BaseMessage
from langchain_core.outputs import ChatGeneration, ChatResult
from pydantic import Field


def say(text: str) -> AIMessage:
    return AIMessage(content=text)


def call(name: str, **args: Any) -> AIMessage:
    return AIMessage(content="", tool_calls=[{"name": name, "args": args, "id": f"c{abs(hash((name, str(args)))) % 10**8}"}])


class Scripted(BaseChatModel):
    """Answers with the next item of the script: a message, or a function of what it was asked."""

    script: list[Any] = Field(default_factory=list)
    asked: list[list[BaseMessage]] = Field(default_factory=list)
    toolsets: list[Any] = Field(default_factory=list)

    @property
    def _llm_type(self) -> str:
        return "scripted"

    def bind_tools(self, tools: Any, **kwargs: Any) -> "Scripted":
        bound = self.model_copy()  # shares the lists below on purpose: one script for the whole run
        object.__setattr__(bound, "script", self.script)
        object.__setattr__(bound, "asked", self.asked)
        bound.toolsets.append([getattr(t, "name", getattr(t, "__name__", str(t))) for t in tools])

        return bound

    def _generate(self, messages: list[BaseMessage], stop: Any = None, run_manager: Any = None, **kw: Any) -> ChatResult:
        self.asked.append(list(messages))

        if not self.script:
            raise AssertionError("the model was asked more times than the test scripted")

        nxt = self.script.pop(0)
        msg = nxt(messages) if callable(nxt) else nxt

        return ChatResult(generations=[ChatGeneration(message=msg)])


class FakeBackend:
    """Backend tools by name; a handler is a value or a function of the arguments."""

    def __init__(self, handlers: dict[str, Any] | None = None) -> None:
        self.handlers = handlers or {}
        self.calls: list[tuple[str, dict[str, Any]]] = []
        self.groups_seen: list[frozenset[str]] = []
        self.specs: list[dict[str, Any]] = []

    async def call(self, tool: str, args: dict[str, Any], tenant: str | None = None,
                   groups: frozenset[str] = frozenset()) -> Any:
        self.calls.append((tool, args))
        self.groups_seen.append(frozenset(groups))
        h = self.handlers.get(tool)

        if h is None:
            raise KeyError(f"the fake backend has no {tool}")

        res = h(args) if callable(h) else h

        if isinstance(res, Exception):
            raise res

        return res

    async def list_tools(self, groups: frozenset[str] = frozenset()) -> list[dict[str, Any]]:
        self.listed = getattr(self, "listed", []) + [frozenset(groups)]

        return self.specs

    def names(self) -> list[str]:
        return [n for n, _ in self.calls]
