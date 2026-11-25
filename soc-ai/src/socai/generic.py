"""The backend tools no composite tool covers, offered to the agent whose area they belong to.

Names go through unchanged except for the dots, which model providers do not accept in a tool name.
"""

from __future__ import annotations

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


def mcp_tools(backend: Backend, specs: list[dict[str, Any]], prefixes: tuple[str, ...], *, tenant: str | None,
              allow_changes: bool, hide: bool = False) -> dict[str, StructuredTool]:
    """One tool for each backend tool under prefixes. Tools that change something need allow_changes."""
    out: dict[str, StructuredTool] = {}

    for spec in specs:
        name = spec["name"]

        if name.startswith(NEVER) or not name.startswith(prefixes):
            continue

        if not spec["read_only"] and not allow_changes:
            continue

        def make(tool: str) -> Any:
            async def run(**kwargs: Any) -> str:
                try:
                    result = await backend.call(tool, kwargs, tenant)

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
