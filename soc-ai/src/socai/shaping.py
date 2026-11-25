"""Makes what a tool returns small enough for a model to use.

The backend's tools return the full documents, with every empty field. A model that is handed a page
of those spends its calls paging through noise instead of concluding. Everything a tool hands back
goes through shape() first: empties dropped, long lists cut with a note saying how many were left, long
text clipped.
"""

from __future__ import annotations

import json
from typing import Any

MAX_ITEMS = 12
MAX_STRING = 400
MAX_CHARS = 6000


def _empty(v: Any) -> bool:
    return v is None or v == "" or v == [] or v == {}


def shape(value: Any, *, max_items: int = MAX_ITEMS, max_string: int = MAX_STRING, depth: int = 0) -> Any:
    if isinstance(value, dict):
        out: dict[str, Any] = {}

        for k, v in value.items():
            s = shape(v, max_items=max_items, max_string=max_string, depth=depth + 1)

            if not _empty(s):
                out[k] = s

        return out

    if isinstance(value, list):
        items = [shape(v, max_items=max_items, max_string=max_string, depth=depth + 1) for v in value[:max_items]]
        items = [i for i in items if not _empty(i)]

        if len(value) > max_items:
            items.append({"_note": f"{len(value) - max_items} more not shown; narrow the query instead of paging"})

        return items

    if isinstance(value, str) and len(value) > max_string:
        return value[:max_string] + f"... [{len(value) - max_string} more characters]"

    return value


def project(doc: dict[str, Any], fields: list[str]) -> dict[str, Any]:
    """Keeps only the dotted paths in fields, in the same nesting the document has."""
    out: dict[str, Any] = {}

    for path in fields:
        parts = path.split(".")
        cur: Any = doc

        for part in parts:
            cur = cur.get(part) if isinstance(cur, dict) else None

            if cur is None:
                break

        if _empty(cur):
            continue

        node = out

        for part in parts[:-1]:
            node = node.setdefault(part, {})

        node[parts[-1]] = cur

    return out


def render(value: Any, *, max_chars: int = MAX_CHARS) -> str:
    """The text a tool result becomes for the model, never longer than max_chars."""
    text = json.dumps(shape(value), ensure_ascii=False, default=str, separators=(",", ":"))

    if len(text) > max_chars:
        text = text[:max_chars] + f"... [cut at {max_chars} characters of {len(text)}; ask for less]"

    return text
