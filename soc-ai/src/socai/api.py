"""The HTTP surface: a health check, the chat that streams every step, and a manual triage."""

from __future__ import annotations

import json
import logging
from typing import Any, AsyncIterator

import asyncio

from fastapi import FastAPI, Header, HTTPException, Request
from fastapi.responses import StreamingResponse
from langchain_core.messages import AIMessage, HumanMessage, ToolMessage
from pydantic import BaseModel, Field

from .config import normalize_groups
from .generic import backend_name
from .runtime import NotConfigured, Runtime

log = logging.getLogger("socai.api")

GENERIC_ERROR = "An error has occurred while processing your request."
OUTPUT_CAP = 1500


class Turn(BaseModel):
    role: str
    content: str


class TaskRequest(BaseModel):
    task: str = Field(min_length=1)
    page: str = ""
    lang: str | None = None
    enabled_groups: list[str] | None = None
    history: list[Turn] = Field(default_factory=list)


class AlertRequest(BaseModel):
    alert_id: str
    tenant: str | None = None


def _sse(event: str, data: dict[str, Any]) -> str:
    """The page reads only the data line and tells the steps apart by its "kind", so the kind goes there too."""
    return f"event: {event}\ndata: {json.dumps({'kind': event, **data}, ensure_ascii=False, default=str)}\n\n"


class Narrator:
    """Turns graph updates into the steps the chat page shows, naming the specialist doing each one."""

    def __init__(self) -> None:
        self.agent = "supervisor"

    def feed(self, namespace: tuple[str, ...], update: dict[str, Any]) -> list[tuple[str, dict[str, Any]]]:
        out: list[tuple[str, dict[str, Any]]] = []
        inside = bool(namespace)

        for node, value in update.items():
            if not isinstance(value, dict):
                continue

            pending = value.get("pending")

            if not inside and node == "route" and pending is not None and pending.action == "delegate":
                self.agent = pending.agent or "supervisor"
                out.append(("tool_call", {"agent": "supervisor", "tool": f"ask:{self.agent}",
                                          "args": {"task": pending.task}}))

            who = self.agent if inside else "supervisor"

            for m in value.get("messages") or []:
                if isinstance(m, AIMessage):
                    for c in m.tool_calls:
                        out.append(("tool_call", {"agent": who, "tool": backend_name(c["name"]), "args": c["args"]}))
                elif isinstance(m, ToolMessage):
                    text = str(m.content)
                    out.append(("tool_result", {"agent": who, "tool": backend_name(m.name or ""),
                                                "output": text[:OUTPUT_CAP], "isError": text.startswith("ERROR:")}))

            if not inside and node == "finish" and value.get("answer"):
                out.append(("final", {"text": value["answer"]}))

        return out


def create_app(rt: Runtime) -> FastAPI:
    app = FastAPI(title="UTMStack SOC AI")

    def authorize(key: str | None) -> None:
        if not rt.settings.internal_key or key != rt.settings.internal_key:
            raise HTTPException(status_code=401, detail="unauthorized")

    @app.get("/health")
    async def health() -> dict[str, str]:
        return {"status": "ok"}

    @app.post("/api/v1/agent/task")
    async def task(req: TaskRequest, x_internal_key: str | None = Header(None),
                   x_tenant_id: str | None = Header(None)) -> StreamingResponse:
        authorize(x_internal_key)

        async def stream() -> AsyncIterator[str]:
            try:
                groups = normalize_groups(req.enabled_groups) if req.enabled_groups is not None else None
                graph = await rt.chat_graph(tenant=x_tenant_id, groups=groups)
                messages = [HumanMessage(t.content) if t.role == "user" else AIMessage(t.content) for t in req.history]
                messages.append(HumanMessage(req.task))
                context = f"The user is currently on: {req.page}" if req.page else ""
                final = False
                narrator = Narrator()

                async for ns, update in graph.astream({"messages": messages, "context": context, "lang": req.lang},
                                                      stream_mode="updates", subgraphs=True):
                    for kind, data in narrator.feed(ns, update):
                        final = final or kind == "final"
                        yield _sse(kind, data)

                if not final:
                    yield _sse("error", {"text": GENERIC_ERROR})
            except NotConfigured as exc:
                yield _sse("error", {"text": str(exc)})
            except Exception:  # noqa: BLE001 - the page gets a generic message, the log gets the cause
                log.exception("the chat failed")
                yield _sse("error", {"text": GENERIC_ERROR})

        return StreamingResponse(stream(), media_type="text/event-stream")

    background: set[asyncio.Task] = set()

    @app.post("/api/v1/analyze", status_code=202)
    async def analyze(request: Request, x_internal_key: str | None = Header(None),
                      x_tenant_id: str | None = Header(None)) -> dict[str, str]:
        """A person asked for this alert to be analysed (the "analyze" button). The backend sends the alert
        and waits only for the acceptance; the assessment is written to the alert when it is done."""
        authorize(x_internal_key)

        try:
            body = await request.json()
        except ValueError:
            body = None

        alert_id = body.get("id") if isinstance(body, dict) else None

        if not alert_id:
            raise HTTPException(status_code=400, detail="Alert ID is required")

        tenant = x_tenant_id or (body.get("tenantId") if isinstance(body, dict) else None)

        async def run() -> None:
            try:
                outcome = await rt.triage(str(alert_id), tenant, force=True)
                log.info("manual analysis of %s: %s", alert_id, outcome)
            except Exception:  # noqa: BLE001
                log.exception("manual analysis of %s failed", alert_id)

        task = asyncio.create_task(run())
        background.add(task)
        task.add_done_callback(background.discard)

        return {"status": "queued", "alertId": str(alert_id), "message": "Alert queued for analysis"}

    @app.post("/api/v1/agent/alert")
    async def alert(req: AlertRequest, x_internal_key: str | None = Header(None)) -> dict[str, str]:
        authorize(x_internal_key)

        return {"alertId": req.alert_id, "outcome": await rt.triage(req.alert_id, req.tenant)}

    return app
