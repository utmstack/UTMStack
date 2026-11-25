"""McpBackend against a real MCP server over HTTP, so a change in the MCP library cannot break it unseen.

The rest of the suite uses a fake backend; this is the only place the real client code runs. It once
broke on a newer release of the library (a renamed function), and nothing noticed until it was deployed.
"""

import socket
import threading
import time

import pytest
import uvicorn
from mcp.server.fastmcp import FastMCP

from socai.backend import McpBackend, ToolError


class Capture:
    """An ASGI wrapper that remembers the headers of every request."""

    def __init__(self, app):
        self.app, self.seen = app, []

    async def __call__(self, scope, receive, send):
        if scope["type"] == "http":
            self.seen.append({k.decode().lower(): v.decode() for k, v in scope["headers"]})

        await self.app(scope, receive, send)


def free_port() -> int:
    with socket.socket() as s:
        s.bind(("127.0.0.1", 0))

        return s.getsockname()[1]


@pytest.fixture
def server():
    mcp = FastMCP("test", stateless_http=True, json_response=True)

    @mcp.tool()
    def alerts_score(alert_id: str) -> dict:
        """Score an alert."""
        return {"score": 81, "alert": alert_id}

    @mcp.tool()
    def broken() -> dict:
        """Always fails."""
        raise ValueError("the backend refused this")

    app = Capture(mcp.streamable_http_app())
    port = free_port()
    srv = uvicorn.Server(uvicorn.Config(app, host="127.0.0.1", port=port, log_level="error"))
    t = threading.Thread(target=srv.run, daemon=True)
    t.start()

    for _ in range(100):
        if srv.started:
            break

        time.sleep(0.05)

    assert srv.started, "the test server did not start"

    yield f"http://127.0.0.1:{port}", app

    srv.should_exit = True
    t.join(5)


# The real backend serves /api/v1/mcp; the test server serves /mcp, so point the client at a base whose
# suffix lands there.
def backend_for(url: str) -> McpBackend:
    b = McpBackend(url, "the-key")
    b._url = url + "/mcp"

    return b


async def test_it_lists_tools_and_calls_one_with_the_headers_the_backend_reads(server):
    url, app = server
    b = backend_for(url)

    tools = await b.list_tools(frozenset({"soar", "alert_actions"}))
    assert {t["name"] for t in tools} >= {"alerts_score", "broken"}
    assert all({"name", "description", "schema", "read_only"} <= set(t) for t in tools)

    assert await b.call("alerts_score", {"alert_id": "a1"}, "t-1", frozenset({"soar"})) == {"score": 81, "alert": "a1"}

    first, call = app.seen[0], app.seen[-1]
    assert first["x-internal-key"] == "the-key" and first["x-agent-groups"] == "alert_actions,soar"
    assert call["x-agent-groups"] == "soar" and call["x-tenant-id"] == "t-1"

    await b.close()


async def test_a_refusal_from_the_backend_reaches_the_model_as_a_tool_error(server):
    url, _ = server
    b = backend_for(url)

    with pytest.raises(ToolError, match="refused"):
        await b.call("broken", {})

    await b.close()


async def test_each_tenant_and_set_of_groups_has_its_own_connection(server):
    url, app = server
    b = backend_for(url)

    await b.call("alerts_score", {"alert_id": "x"}, "tenant-a", frozenset())
    await b.call("alerts_score", {"alert_id": "x"}, "tenant-b", frozenset())
    await b.call("alerts_score", {"alert_id": "x"}, "tenant-a", frozenset({"soar"}))

    scopes = {(h.get("x-tenant-id"), h.get("x-agent-groups")) for h in app.seen}
    assert scopes == {("tenant-a", ""), ("tenant-b", ""), ("tenant-a", "soar")}

    await b.close()


async def test_an_unreachable_backend_is_a_tool_error_not_a_crash():
    b = McpBackend("http://127.0.0.1:1", "k")

    with pytest.raises(ToolError, match="could not be reached"):
        await b.call("alerts_score", {"alert_id": "x"})

    await b.close()


async def test_a_connection_opened_by_one_task_can_be_used_and_closed_by_others(server):
    """The chat and the alert worker call from different tasks, and the service closes from yet another."""
    import asyncio

    url, _ = server
    b = backend_for(url)

    first = await asyncio.create_task(b.call("alerts_score", {"alert_id": "a"}, "t", frozenset()))
    second = await asyncio.create_task(b.call("alerts_score", {"alert_id": "b"}, "t", frozenset()))

    assert first["alert"] == "a" and second["alert"] == "b"

    await asyncio.create_task(b.close())  # closed by a task that never opened it: this used to raise


@pytest.mark.parametrize("json_response", [True, False])
def test_parallel_calls_share_one_connection(json_response):
    import asyncio

    mcp = FastMCP("test", stateless_http=True, json_response=json_response)

    @mcp.tool()
    async def slow(n: int) -> dict:
        """Answer after a moment."""
        await asyncio.sleep(0.2)

        return {"n": n}

    port = free_port()
    srv = uvicorn.Server(uvicorn.Config(mcp.streamable_http_app(), host="127.0.0.1", port=port, log_level="error"))
    t = threading.Thread(target=srv.run, daemon=True)
    t.start()

    for _ in range(100):
        if srv.started:
            break

        time.sleep(0.05)

    try:
        async def go():
            b = backend_for(f"http://127.0.0.1:{port}")

            try:
                return await asyncio.wait_for(asyncio.gather(*(b.call("slow", {"n": i}) for i in range(4))), 15)
            finally:
                await b.close()

        assert sorted(r["n"] for r in asyncio.run(go())) == [0, 1, 2, 3]
    finally:
        srv.should_exit = True
        t.join(5)


def test_a_bad_answer_does_not_take_the_other_calls_down():
    import asyncio

    from mcp.types import CallToolResult

    class Session:
        async def call_tool(self, name, args):
            if name == "slow":
                await asyncio.sleep(0.3)

                return CallToolResult(content=[], structuredContent={"ok": True})

            return CallToolResult.model_validate({"content": [], "structuredContent": [1, 2]})

    b = McpBackend("http://x", "k")
    sessions = []

    async def fake_session(scope):
        sessions.append(scope)

        return Session()

    async def go():
        b._session_for = fake_session
        slow = asyncio.ensure_future(b.call("slow", {}))

        with pytest.raises(ToolError, match="not valid"):
            await b.call("bad", {})

        return await asyncio.wait_for(slow, 5)

    assert asyncio.run(go()) == {"ok": True}
    assert len(sessions) == 2, "a bad answer must not reconnect"


def test_a_call_that_never_answers_gives_up():
    import asyncio

    class Hung:
        async def call_tool(self, name, args):
            await asyncio.sleep(60)

    b = McpBackend("http://x", "k")
    b.call_timeout = 0.2

    async def fake_session(scope):
        return Hung()

    b._session_for = fake_session

    with pytest.raises(ToolError, match="did not answer"):
        asyncio.run(b.call("anything", {}))
