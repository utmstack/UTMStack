import httpx

from socai.api import create_app
from socai.config import Settings
from socai.runtime import Runtime
from fakes import FakeBackend, Scripted, call, say

KEY = {"X-Internal-Key": "k"}


def client(model, backend=None):
    rt = Runtime(Settings(internal_key="k"), backend or FakeBackend(), model)

    return httpx.AsyncClient(transport=httpx.ASGITransport(app=create_app(rt)), base_url="http://t")


async def test_health_needs_no_key():
    async with client(Scripted()) as c:
        assert (await c.get("/health")).json() == {"status": "ok"}


async def test_the_chat_refuses_a_caller_without_the_key():
    async with client(Scripted()) as c:
        assert (await c.post("/api/v1/agent/task", json={"task": "hi"})).status_code == 401
        assert (await c.post("/api/v1/agent/task", json={"task": "hi"}, headers={"X-Internal-Key": "x"})).status_code == 401


async def test_the_chat_streams_each_step_and_the_final_answer():
    model = Scripted(script=[
        call("Route", action="delegate", agent="alerts", task="how many are open"),
        call("find_alerts", status="open"),
        say("There are 2 open alerts."),
        call("Route", action="finish", answer="2 alerts are open."),
    ])
    backend = FakeBackend({"store.search": {"hits": [{"id": "a", "name": "n"}], "total": 2}})

    async with client(model, backend) as c:
        r = await c.post("/api/v1/agent/task", json={"task": "open alerts?"}, headers=KEY)

    body = r.text
    assert r.status_code == 200 and "text/event-stream" in r.headers["content-type"]
    assert "event: tool_call" in body and "ask:alerts" in body
    assert '"kind": "tool_call"' in body and '"kind": "final"' in body, "the page tells steps apart by data.kind"
    assert '"agent": "alerts", "tool": "find_alerts"' in body, "each step names the specialist that took it"
    assert "event: tool_result" in body
    assert body.rstrip().endswith('"text": "2 alerts are open."}'), body[-200:]


async def test_a_failure_reaches_the_page_as_a_generic_error():
    model = Scripted(script=[])  # asking it raises

    async with client(model) as c:
        r = await c.post("/api/v1/agent/task", json={"task": "hi"}, headers=KEY)

    assert "event: error" in r.text and "An error has occurred" in r.text and "scripted" not in r.text


async def test_manual_triage_reports_what_happened():
    backend = FakeBackend({"store.search": {"hits": [], "total": 0}, "alerts.score": {"score": 1}})

    async with client(Scripted(), backend) as c:
        r = await c.post("/api/v1/agent/alert", json={"alert_id": "zz"}, headers=KEY)

    assert r.json()["outcome"].startswith("failed")


async def test_the_analyze_button_is_accepted_at_once_and_the_alert_is_assessed_again():
    seen = []

    class Rt(Runtime):
        async def triage(self, alert_id, tenant, facts=None, force=False):
            seen.append((alert_id, tenant, force))

            return "written"

    rt = Rt(Settings(internal_key="k"), FakeBackend(), Scripted())

    async with httpx.AsyncClient(transport=httpx.ASGITransport(app=create_app(rt)), base_url="http://t") as c:
        r = await c.post("/api/v1/analyze", json={"id": "a1", "tenantId": "t-body"}, headers={**KEY, "X-Tenant-Id": "t1"})

        assert r.status_code == 202 and r.json()["status"] == "queued" and r.json()["alertId"] == "a1"

        import asyncio
        await asyncio.sleep(0.05)

    assert seen == [("a1", "t1", True)], "the tenant of the request wins, and a person's request is never skipped"


async def test_the_analyze_button_refuses_what_is_not_an_alert_or_a_caller_without_the_key():
    async with client(Scripted()) as c:
        assert (await c.post("/api/v1/analyze", json={"name": "no id"}, headers=KEY)).status_code == 400
        assert (await c.post("/api/v1/analyze", content=b"not json", headers=KEY)).status_code == 400
        assert (await c.post("/api/v1/analyze", json={"id": "a1"})).status_code == 401
