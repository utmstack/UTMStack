"""The tools the agents get.

They are not one wrapper per backend call. A model that is given a hundred small tools makes a hundred
small calls and runs out of steps before it has said anything; so the tools here answer a question:
"what is this alert and what is around it" is one call that returns a short dossier, not five calls
whose raw output the model has to join itself.

Anything that changes something is marked with ACTION and only offered when an administrator allowed
that kind of change.
"""

from __future__ import annotations

import asyncio
import re
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from typing import Any, Awaitable, Callable

from langchain_core.tools import StructuredTool
from pydantic import BaseModel, Field

from .backend import Backend, ToolError
from .anonymize import hide_field_values, hide_people
from .shaping import project, render

MARK = "[AI SOC Agent]"

ALERT_FIELDS = [
    "id", "name", "severity", "status", "category", "technique", "@timestamp", "dataSource", "dataType", "description",
    "adversary.ip", "adversary.host", "adversary.user", "adversary.port",
    "target.ip", "target.host", "target.user", "target.port",
    "tags", "notes", "impactScore", "echoes", "tenantId", "parentId",
]

RELATED_FIELDS = ["id", "name", "severity", "status", "@timestamp", "adversary.ip", "target.ip", "target.host", "tags"]

_SINCE = re.compile(r"^(\d+)([mhdw])$")
_UNITS = {"m": "minutes", "h": "hours", "d": "days", "w": "weeks"}

# An alert's status is stored as text, but the tool that changes it takes a number.
STATUS_TEXT = {"open": "Open", "in_review": "In review", "closed": "Completed"}
STATUS_CODE = {"open": 0, "in_review": 1, "closed": 2}


_UNITS = {"m": 60, "h": 3600, "d": 86400, "w": 604800}
_COMPOUND = re.compile(r"^(?:\d+[mhdw])+$")
_PART = re.compile(r"(\d+)([mhdw])")


def _now() -> datetime:
    return datetime.now(timezone.utc)


def _iso(t: datetime) -> str:
    return t.astimezone(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def since_filter(since: str | None) -> list[dict[str, Any]]:
    """A window as a filter on the timestamp: "30m", "24h", "7d", "2w", "1h30m", "today", or a date/time.

    The bounds go out as absolute times. The backend ignores a bound it cannot read, which would leave the
    query without any limit; an absolute time is read the same everywhere."""
    if not since:
        return []

    text = since.strip().lower()
    now = _now()

    if text == "today":
        start = now.replace(hour=0, minute=0, second=0, microsecond=0)
    elif _COMPOUND.match(text):
        start = now - timedelta(seconds=sum(int(n) * _UNITS[u] for n, u in _PART.findall(text)))
    else:
        start = None

        for layout in ("%Y-%m-%dT%H:%M:%SZ", "%Y-%m-%dT%H:%M:%S", "%Y-%m-%d %H:%M:%S", "%Y-%m-%d"):
            try:
                start = datetime.strptime(since.strip().rstrip("Z") + ("Z" if layout.endswith("Z") else ""), layout).replace(tzinfo=timezone.utc)
                break
            except ValueError:
                continue

        if start is None:
            raise ToolError(f'since must look like "30m", "24h", "7d", "2w", "today" or a date such as "2026-11-25", got {since!r}')

    return [{"field": "@timestamp", "operator": "IS_BETWEEN", "value": [_iso(start), _iso(now)]}]


def eq(field: str, value: Any) -> dict[str, Any]:
    return {"field": field, "operator": "IS", "value": value}


def _hits(res: Any) -> list[dict[str, Any]]:
    if isinstance(res, dict):
        return list(res.get("hits") or res.get("items") or [])

    return list(res or [])


@dataclass
class Toolbox:
    """The tools over one backend, for one tenant."""

    backend: Backend
    tenant: str | None = None
    # The capability groups this toolbox asks the backend to turn on; the backend shows nothing outside
    # them, and the changes below also refuse to run without theirs.
    groups: frozenset[str] = frozenset()
    # Triage runs on its own and must not hand people's names to a model.
    hide_people: bool = False

    async def _call(self, tool: str, args: dict[str, Any]) -> Any:
        return await self.backend.call(tool, args, self.tenant, self.groups)

    # --- reading --------------------------------------------------------------------------------

    async def alert_dossier(self, alert_id: str) -> dict[str, Any]:
        """Everything needed to judge one alert, in one call."""
        found, score = await asyncio.gather(
            self._call("store.search", {"dataset": "alerts", "filters": [eq("id", alert_id)], "size": 1}),
            self._call("alerts.score", {"alert_id": alert_id}),
            return_exceptions=True,
        )

        hits = [] if isinstance(found, BaseException) else _hits(found)

        if not hits:
            raise ToolError(f"no alert with id {alert_id}")

        alert = hits[0]
        out: dict[str, Any] = {"alert": project(alert, ALERT_FIELDS)}

        if not isinstance(score, BaseException):
            out["score"] = score

        # Who else was involved and how often this has happened, answered together.
        adversary = (alert.get("adversary") or {}).get("ip")
        target = (alert.get("target") or {}).get("ip")
        name = alert.get("name")

        queries: dict[str, Awaitable[Any]] = {}

        for label, ip_field, ip in (("same_adversary", "adversary.ip", adversary), ("same_target", "target.ip", target)):
            if ip:
                queries[label] = self._call(
                    "store.search",
                    {"dataset": "alerts", "filters": [eq(ip_field, ip), *since_filter("7d")], "size": 8,
                     "sort_by": "@timestamp", "sort_order": "desc"},
                )

        if name:
            # The search answers with the total; one row is enough to get it.
            queries["same_name_30d"] = self._call(
                "store.search", {"dataset": "alerts", "filters": [eq("name", name), *since_filter("30d")], "size": 1}
            )

        results = await asyncio.gather(*queries.values(), return_exceptions=True)

        for label, res in zip(queries, results):
            if isinstance(res, BaseException):
                continue

            if label == "same_name_30d":
                out[label] = res.get("total") if isinstance(res, dict) else res
            else:
                out[label] = [project(h, RELATED_FIELDS) for h in _hits(res) if h.get("id") != alert_id]

        return out

    async def find_alerts(
        self,
        severity: str | None = None,
        status: str | None = None,
        name_contains: str | None = None,
        ip: str | None = None,
        host: str | None = None,
        since: str | None = None,
        limit: int = 10,
    ) -> dict[str, Any]:
        # Without a window it counts every alert: an implicit one answered "how many open alerts" with the
        # last day's, and the person could not tell.
        filters = since_filter(since)

        if severity:
            filters.append(eq("severity", severity))

        if status:
            if status not in STATUS_TEXT:
                raise ToolError(f"status must be one of {sorted(STATUS_TEXT)}")

            filters.append(eq("status", STATUS_TEXT[status]))

        if name_contains:
            filters.append({"field": "name", "operator": "CONTAIN", "value": name_contains})

        if ip:
            filters.append({"field": "adversary.ip", "operator": "IS_ONE_OF", "value": [ip]})

        if host:
            filters.append({"field": "target.host", "operator": "IS", "value": host})

        res = await self._call(
            "store.search",
            {"dataset": "alerts", "filters": filters, "size": min(max(limit, 1), 25), "sort_by": "@timestamp",
             "sort_order": "desc"},
        )
        total = res.get("total") if isinstance(res, dict) else None

        return {"total": total, "window": since or "all time", "alerts": [project(h, RELATED_FIELDS) for h in _hits(res)]}

    async def top_values(self, dataset: str, field: str, since: str | None = "24h", filters: list[dict] | None = None,
                         top: int = 10) -> Any:
        res = await self._call(
            "store.property_values",
            {"dataset": dataset, "field": field, "filters": [*since_filter(since), *(filters or [])], "top": min(top, 25)},
        )

        return hide_field_values(field, res) if self.hide_people else res

    async def search_logs(self, data_type: str | None = None, text: str | None = None, ip: str | None = None,
                          host: str | None = None, user: str | None = None, since: str = "1h",
                          limit: int = 8) -> dict[str, Any]:
        filters = since_filter(since)

        if ip:
            filters.append({"field": "origin.ip", "operator": "IS", "value": ip})

        if host:
            filters.append({"field": "origin.host", "operator": "IS", "value": host})

        if user:
            filters.append({"field": "origin.user", "operator": "IS", "value": user})

        if text:
            filters.append({"field": "raw", "operator": "CONTAIN", "value": text})

        args: dict[str, Any] = {"dataset": "logs", "filters": filters, "size": min(max(limit, 1), 15),
                                "sort_by": "@timestamp", "sort_order": "desc"}

        if data_type:
            args["data_type"] = data_type

        res = await self._call("store.search", args)
        keep = ["@timestamp", "dataType", "dataSource", "origin.ip", "origin.host", "origin.user", "target.ip",
                "target.host", "action", "actionResult", "message"]

        return {"total": res.get("total") if isinstance(res, dict) else None,
                "logs": [project(h, keep) for h in _hits(res)]}

    async def incident_summary(self, incident_id: str) -> dict[str, Any]:
        inc, alerts, notes = await asyncio.gather(
            self._call("incidents.get", {"id": incident_id}),
            self._call("incident_alerts.list", {"incident_id": incident_id, "size": 10}),
            self._call("incident_notes.list", {"incident_id": incident_id, "size": 5}),
            return_exceptions=True,
        )

        return {k: v for k, v in (("incident", inc), ("alerts", alerts), ("notes", notes))
                if not isinstance(v, BaseException)}

    async def list_incidents(self, status: str | None = None, limit: int = 10) -> Any:
        args: dict[str, Any] = {"size": min(max(limit, 1), 25)}

        if status:
            args["status"] = status

        return await self._call("incidents.list", args)

    # --- changing things (ACTION) ---------------------------------------------------------------

    def _require(self, group: str) -> None:
        if group not in self.groups:
            raise ToolError(f'changes to {group} are not enabled for this agent; say what you would do instead')

    async def add_alert_note(self, alert_id: str, note: str) -> Any:
        # The backend keeps the analyst's own text when what it is given opens with the assessment marker,
        # and replaces only an earlier assessment. Anything else would replace the notes, so keep them.
        if note.startswith(MARK):
            return await self._call("alerts.update_notes", {"alert_id": alert_id, "notes": note})

        current = await self.alert_dossier(alert_id)
        existing = (current.get("alert") or {}).get("notes") or ""
        joined = f"{existing}\n{note}".strip() if existing and note not in existing else (existing or note)

        return await self._call("alerts.update_notes", {"alert_id": alert_id, "notes": joined})

    async def set_alert_status(self, alert_ids: list[str], status: str, observation: str = "") -> Any:
        self._require("alert_actions")

        if status not in STATUS_CODE:
            raise ToolError(f"status must be one of {sorted(STATUS_CODE)}")

        return await self._call("alerts.update_status",
                                {"alert_ids": alert_ids, "status": STATUS_CODE[status], "status_observation": observation})

    async def tag_alerts(self, alert_ids: list[str], tags: list[str]) -> Any:
        self._require("alert_actions")

        return await self._call("alerts.update_tags", {"alert_ids": alert_ids, "tags": tags})

    async def create_incident(self, alert_ids: list[str], name: str) -> Any:
        self._require("incident_actions")

        return await self._call("alerts.convert_to_incident",
                                {"alert_ids": alert_ids, "incident_name": name, "incident_id": 0,
                                 "incident_source": "soc-ai"})


# --- the tools as the models see them -----------------------------------------------------------


class _AlertId(BaseModel):
    alert_id: str = Field(description="The alert's id")


class _FindAlerts(BaseModel):
    severity: str | None = Field(None, description="low, medium or high")
    status: str | None = Field(None, description="open, in_review or closed")
    name_contains: str | None = None
    ip: str | None = Field(None, description="Adversary IP address")
    host: str | None = Field(None, description="Target host name")
    since: str | None = Field(None, description='Window such as "30m", "24h", "7d", "today". Leave it out to count every alert, '
                              'and pass it whenever the question names a period.')
    limit: int = Field(10, description="At most 25")


class _TopValues(BaseModel):
    dataset: str = Field(description="alerts or logs")
    field: str = Field(description=(
        'Exact field path. Logs: "origin.user", "origin.ip", "origin.host", "target.ip", "dataType", "action". '
        'Alerts: "adversary.ip", "target.user", "name", "severity", "status". A bare "user" or "ip" is not a field; '
        'use the dotted path.'))
    since: str | None = "24h"
    top: int = 10


class _SearchLogs(BaseModel):
    data_type: str | None = None
    text: str | None = Field(None, description="Text the message must contain")
    ip: str | None = Field(None, description="Origin IP")
    host: str | None = Field(None, description="Origin host")
    user: str | None = Field(None, description="Origin user")
    since: str = "1h"
    limit: int = 8


class _IncidentId(BaseModel):
    incident_id: str = Field(description="The incident's UUID, as listed by list_incidents. It is not a number.")


class _ListIncidents(BaseModel):
    status: str | None = Field(None, description="Open, In review or Closed")
    limit: int = 10


class _Note(BaseModel):
    alert_id: str
    note: str


class _Status(BaseModel):
    alert_ids: list[str]
    status: str = Field(description="open, in_review or closed")
    observation: str = ""


class _Tags(BaseModel):
    alert_ids: list[str]
    tags: list[str]


class _Incident(BaseModel):
    alert_ids: list[str]
    name: str


def _tool(name: str, description: str, schema: type[BaseModel], fn: Callable[..., Awaitable[Any]],
          hide: bool = False) -> StructuredTool:
    async def run(**kwargs: Any) -> str:
        try:
            result = await fn(**kwargs)

            if hide:
                result, hidden = hide_people(result)

                if hidden and isinstance(result, dict):
                    result["anonymizedFields"] = sorted(set(hidden))

            return render(result)
        except ToolError as exc:
            return f"ERROR: {exc}"

    return StructuredTool.from_function(coroutine=run, name=name, description=description, args_schema=schema)


def read_tools(tb: Toolbox) -> dict[str, StructuredTool]:
    h = tb.hide_people

    return {
        "get_alert_dossier": _tool(
            "get_alert_dossier",
            "Everything about one alert in a single call: its fields, its risk score with the breakdown, the other "
            "alerts from the same adversary and the same target in the last 7 days, and how many alerts with this "
            "name there were in 30 days. Call this first; it usually makes other lookups unnecessary.",
            _AlertId, lambda alert_id: tb.alert_dossier(alert_id), h),
        "find_alerts": _tool(
            "find_alerts", "Find alerts by simple criteria. Returns the total (over the window it names) and up to 25 short rows.",
            _FindAlerts, tb.find_alerts, h),
        "top_values": _tool(
            "top_values", "The most common values of one field, with counts. Use it to see what stands out.",
            _TopValues, tb.top_values, h),
        "search_logs": _tool(
            "search_logs", "Recent log events for an address, host or user. Returns a total and a few short rows.",
            _SearchLogs, tb.search_logs, h),
        "get_incident": _tool(
            "get_incident", "An incident with its alerts and latest notes, in one call.", _IncidentId,
            lambda incident_id: tb.incident_summary(incident_id), h),
        "list_incidents": _tool("list_incidents", "List incidents. Each one has an id that is a UUID.", _ListIncidents, tb.list_incidents, h),
    }


def action_tools(tb: Toolbox) -> dict[str, StructuredTool]:
    return {
        "add_alert_note": _tool("add_alert_note", "ACTION: add a note to an alert.", _Note, tb.add_alert_note),
        "set_alert_status": _tool(
            "set_alert_status", "ACTION: change the status of alerts. Only when the evidence is clear.", _Status,
            tb.set_alert_status),
        "tag_alerts": _tool("tag_alerts", "ACTION: tag alerts.", _Tags, tb.tag_alerts),
        "create_incident": _tool(
            "create_incident", "ACTION: open an incident from alerts. Only when they are really one incident.",
            _Incident, tb.create_incident),
    }
