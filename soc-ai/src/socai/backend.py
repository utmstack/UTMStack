"""The backend's tools, over its MCP endpoint.

The agents never build HTTP calls to the backend: every operation is one of the tools it publishes at
/api/v1/mcp, so permissions, tenancy and validation stay where they are.
"""

from __future__ import annotations

import asyncio
import json
from typing import Any, Protocol


class ToolError(Exception):
    """The backend refused or failed a tool call; the message is meant for the model."""


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

OPEN_TIMEOUT = 30.0
CLOSE_TIMEOUT = 5.0


class _Connection:
    """One open MCP session and the task that owns it.

    The library keeps a connection open with a task group that has to be entered and left by the same
    task, and calls come from whichever task asked. So a task of its own opens the connection, holds it,
    and is the one to close it when told to; callers only use the session it publishes."""

    def __init__(self) -> None:
        self.session: Any = None
        self.task: asyncio.Task | None = None
        self.stop = asyncio.Event()

    @property
    def alive(self) -> bool:
        return self.session is not None and self.task is not None and not self.task.done()


class McpBackend:
    """One MCP session for each tenant and set of groups, reopened once if it drops.

    Who the tenant is and which groups are on travel in the headers of the connection, so each
    combination needs its own."""

    def __init__(self, base_url: str, internal_key: str) -> None:
        self._url = base_url.rstrip("/") + "/api/v1/mcp"
        self._key = internal_key
        self._lock = asyncio.Lock()
        self._conns: dict[Scope, _Connection] = {}

    def _headers(self, scope: Scope) -> dict[str, str]:
        tenant, groups = scope
        headers = {"X-Internal-Key": self._key, "X-Agent-Groups": ",".join(sorted(groups))}

        if tenant:
            headers["X-Tenant-Id"] = tenant

        return headers

    async def _hold(self, scope: Scope, conn: _Connection, ready: asyncio.Future) -> None:
        import httpx
        from mcp import ClientSession
        from mcp.client.streamable_http import streamable_http_client

        try:
            async with httpx.AsyncClient(headers=self._headers(scope), timeout=90) as client:
                async with streamable_http_client(self._url, http_client=client) as (read, write, *_):
                    async with ClientSession(read, write) as session:
                        await session.initialize()
                        conn.session = session
                        ready.set_result(None)
                        await conn.stop.wait()
        except BaseException as exc:  # noqa: BLE001 - whoever waits for the connection needs to hear why
            if not ready.done():
                ready.set_exception(exc if isinstance(exc, Exception) else RuntimeError(repr(exc)))

            if not isinstance(exc, Exception):
                raise
        finally:
            conn.session = None

    async def _connect(self, scope: Scope) -> _Connection:
        conn = self._conns.get(scope)

        if conn is not None and conn.alive:
            return conn

        if conn is not None:
            await self._shutdown(scope)

        conn = _Connection()
        ready: asyncio.Future = asyncio.get_running_loop().create_future()
        conn.task = asyncio.create_task(self._hold(scope, conn, ready))

        try:
            await asyncio.wait_for(ready, OPEN_TIMEOUT)
        except BaseException:
            conn.stop.set()
            await asyncio.gather(conn.task, return_exceptions=True)

            raise

        self._conns[scope] = conn

        return conn

    async def _shutdown(self, scope: Scope) -> None:
        conn = self._conns.pop(scope, None)

        if conn is None or conn.task is None:
            return

        conn.stop.set()

        try:
            await asyncio.wait_for(asyncio.gather(conn.task, return_exceptions=True), CLOSE_TIMEOUT)
        except asyncio.TimeoutError:
            conn.task.cancel()

    async def _session_for(self, scope: Scope) -> Any:
        async with self._lock:
            return (await self._connect(scope)).session

    async def call(self, tool: str, args: dict[str, Any], tenant: str | None = None,
                   groups: frozenset[str] = frozenset()) -> Any:
        scope: Scope = (tenant, frozenset(groups))

        for attempt in (1, 2):
            try:
                session = await self._session_for(scope)

                return _parse(await session.call_tool(tool, args))
            except ToolError:
                raise
            except Exception as exc:  # noqa: BLE001
                if attempt == 2:
                    raise ToolError(f"the backend could not be reached: {exc}") from exc

                async with self._lock:
                    await self._shutdown(scope)

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
            for scope in list(self._conns):
                await self._shutdown(scope)
