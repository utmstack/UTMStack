"""The backend's tools, over its MCP endpoint.

The agents never build HTTP calls to the backend: every operation is one of the tools it publishes at
/api/v1/mcp, so permissions, tenancy and validation stay where they are.
"""

from __future__ import annotations

import asyncio
import json
from contextlib import AsyncExitStack
from typing import Any, Protocol


class ToolError(Exception):
    """The backend refused or failed a tool call; the message is meant for the model."""


Groups = frozenset  # the capability groups a request asks the backend to turn on


class Backend(Protocol):
    async def call(self, tool: str, args: dict[str, Any], tenant: str | None = None,
                   groups: frozenset[str] = frozenset()) -> Any: ...

    async def list_tools(self, groups: frozenset[str] = frozenset()) -> list[dict[str, Any]]: ...


def _parse(result: Any) -> Any:
    structured = getattr(result, "structuredContent", None)

    if structured is not None:
        value = structured.get("result", structured) if isinstance(structured, dict) else structured
    else:
        text = "".join(getattr(c, "text", "") for c in (result.content or []))

        try:
            value = json.loads(text)
        except json.JSONDecodeError:
            value = text

    if getattr(result, "isError", False):
        raise ToolError(value if isinstance(value, str) else json.dumps(value))

    return value


Scope = tuple[str | None, frozenset[str]]


class McpBackend:
    """One MCP session for each tenant and set of groups, reopened once if it drops.

    Who the tenant is and which groups are on travel in the headers of the connection, so each
    combination needs its own."""

    def __init__(self, base_url: str, internal_key: str) -> None:
        self._url = base_url.rstrip("/") + "/api/v1/mcp"
        self._key = internal_key
        self._lock = asyncio.Lock()
        self._stacks: dict[Scope, AsyncExitStack] = {}
        self._sessions: dict[Scope, Any] = {}

    async def _open(self, scope: Scope) -> Any:
        from mcp import ClientSession
        from mcp.client.streamable_http import streamablehttp_client

        tenant, groups = scope
        headers = {"X-Internal-Key": self._key, "X-Agent-Groups": ",".join(sorted(groups))}

        if tenant:
            headers["X-Tenant-Id"] = tenant

        stack = AsyncExitStack()
        read, write, _ = await stack.enter_async_context(streamablehttp_client(self._url, headers=headers, timeout=90))
        session = await stack.enter_async_context(ClientSession(read, write))
        await session.initialize()
        self._stacks[scope], self._sessions[scope] = stack, session

        return session

    async def _close(self, scope: Scope) -> None:
        stack = self._stacks.pop(scope, None)
        self._sessions.pop(scope, None)

        if stack is not None:
            try:
                await stack.aclose()
            except Exception:  # noqa: BLE001 - a dead session is what we are replacing
                pass

    async def _session_for(self, scope: Scope) -> Any:
        async with self._lock:
            if scope not in self._sessions:
                await self._open(scope)

            return self._sessions[scope]

    async def call(self, tool: str, args: dict[str, Any], tenant: str | None = None,
                   groups: frozenset[str] = frozenset()) -> Any:
        scope: Scope = (tenant, frozenset(groups))

        for attempt in (1, 2):
            session = await self._session_for(scope)

            try:
                return _parse(await session.call_tool(tool, args))
            except ToolError:
                raise
            except Exception as exc:  # noqa: BLE001
                if attempt == 2:
                    raise ToolError(f"the backend could not be reached: {exc}") from exc

                async with self._lock:
                    await self._close(scope)

        raise ToolError("unreachable")

    async def list_tools(self, groups: frozenset[str] = frozenset()) -> list[dict[str, Any]]:
        session = await self._session_for((None, frozenset(groups)))
        res = await session.list_tools()

        return [
            {
                "name": t.name,
                "description": t.description or "",
                "schema": t.inputSchema or {"type": "object", "properties": {}},
                "read_only": bool(getattr(getattr(t, "annotations", None), "readOnlyHint", False)),
            }
            for t in res.tools
        ]

    async def close(self) -> None:
        async with self._lock:
            for scope in list(self._stacks):
                await self._close(scope)
