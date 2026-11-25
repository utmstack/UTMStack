"""The backend tools no composite tool covers, offered to the agent whose area they belong to.

Names go through unchanged except for the dots, which model providers do not accept in a tool name.
"""

from __future__ import annotations

import json
import re
from typing import Any

from langchain_core.tools import StructuredTool

from .backend import Backend, ToolError
from .anonymize import hide_people
from .shaping import render

# Never offered to a model, whatever the agent: they manage who may do what, not the SIEM's work.
NEVER = ("iam.", "tenants.", "billing.", "branding.", "config.", "audit.", "socai.", "notifications.")


def wire_name(name: str) -> str:
    return name.replace(".", "__")


def backend_name(name: str) -> str:
    return name.replace("__", ".")


_EMAIL = re.compile(r"[\w.+-]+@[\w-]+(\.[\w-]+)+")


def _widget_key(args: dict[str, Any]) -> str | None:
    """What makes a widget the same widget: the dashboard it is on and what it asks of which data."""
    try:
        spec = json.loads(args["spec"]) if isinstance(args.get("spec"), str) else args.get("spec") or {}
    except ValueError:
        return None

    # How many rows, or what the title says, does not make a different question; the window, the filter and
    # what a table shows or sorts by do.
    parts = [args.get("dashboard_id")] + [json.dumps(spec.get(k), sort_keys=True) for k in
                                          ("dataset", "chart", "dimension", "filter", "range", "interval", "columns", "sort")]

    return json.dumps(parts)


def check_before(tool: str, args: dict[str, Any], made: dict[str, str]) -> tuple[dict[str, Any], str | None]:
    """What this agent will not do however the model asks, decided here and not left to a prompt.

    Returns the arguments to send and, when the call must not go out at all, the answer to give the model."""
    args = dict(args)

    # An enabled rule runs on real alerts. Creating one switched on is the model's call, not the user's: it
    # is made inactive, and turning it on is its own tool.
    if tool == "soar.rule.create" and args.get("active"):
        args["active"] = False

    if tool in ("correlation_rule.create", "correlation_rule.update") and args.get("ruleActive"):
        args["ruleActive"] = False

    if tool == "soar.rule.create":
        for name, node in (args.get("nodes") or {}).items():
            if (node or {}).get("executor") == "mail":
                to = str(((node or {}).get("params") or {}).get("to", ""))

                if not _EMAIL.search(to):
                    return args, (f"ERROR: node {name!r} sends mail but has no recipient address. Ask the user who should "
                                  "receive it, and create the rule once you have an address. Do not make one up or use a "
                                  "placeholder.")

    if tool == "visualizations.create":
        key = _widget_key(args)

        if key and key in made:
            return args, (f"ERROR: this dashboard already has a widget that asks the same ({made[key]}). It was not created "
                          "again. If a call timed out earlier, look with visualizations__list before trying again.")

    return args, None


def mcp_tools(backend: Backend, specs: list[dict[str, Any]], prefixes: tuple[str, ...], *, tenant: str | None,
              allow_changes: bool, hide: bool = False, groups: frozenset[str] = frozenset()) -> dict[str, StructuredTool]:
    """One tool for each backend tool under prefixes. Tools that change something need allow_changes.

    groups are the capability groups the call goes out with. The backend answers only for the tools of the
    groups a request turned on, so a call without them reaches the core tools and nothing else: every
    dashboard, SOAR, compliance, data source and rule tool would be "not available"."""
    out: dict[str, StructuredTool] = {}
    made: dict[str, str] = {}  # widgets created in this request, so a retry does not make a copy

    for spec in specs:
        name = spec["name"]

        if name.startswith(NEVER) or not name.startswith(prefixes):
            continue

        if not spec["read_only"] and not allow_changes:
            continue

        def make(tool: str) -> Any:
            async def run(**kwargs: Any) -> str:
                kwargs, refused = check_before(tool, kwargs, made)

                if refused:
                    return refused

                try:
                    result = await backend.call(tool, kwargs, tenant, groups)

                    if tool == "visualizations.create" and (key := _widget_key(kwargs)):
                        made[key] = str(kwargs.get("title") or "an earlier widget")

                    if hide:
                        result, _ = hide_people(result)

                    return render(result)
                except ToolError as exc:
                    return f"ERROR: {exc}"

            return run

        title = wire_name(name)
        prefix = "" if spec["read_only"] else "ACTION: "
        out[title] = StructuredTool.from_function(
            coroutine=make(name), name=title, description=prefix + (spec["description"] or name)[:600],
            args_schema=spec["schema"])

    return out
