"""Hides what a model should not see: the names and addresses of people.

It works on anything a tool returns, so it covers the alert and everything the agent looks up around
it. What was hidden is listed, so the model knows not to read anything into the placeholders.
"""

from __future__ import annotations

import re
from typing import Any

FAKE_USER = "John Doe"
FAKE_EMAIL = "jhondoe@gmail.com"

_EMAIL = re.compile(r"([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})")
_PERSON_KEYS = {"user": FAKE_USER, "username": FAKE_USER, "email": FAKE_EMAIL}


def hide_people(value: Any, _path: str = "", _hidden: list[str] | None = None) -> tuple[Any, list[str]]:
    hidden: list[str] = [] if _hidden is None else _hidden

    if isinstance(value, dict):
        out: dict[str, Any] = {}

        for k, v in value.items():
            here = f"{_path}.{k}" if _path else k

            if k in _PERSON_KEYS and isinstance(v, str) and v:
                out[k] = _PERSON_KEYS[k]
                hidden.append(here)
            else:
                out[k], _ = hide_people(v, here, hidden)

        return out, hidden

    if isinstance(value, list):
        return [hide_people(v, _path, hidden)[0] for v in value], hidden

    if isinstance(value, str):
        cleaned = _EMAIL.sub(FAKE_EMAIL, value)

        if cleaned != value:
            hidden.append(_path)

        return cleaned, hidden

    return value, hidden
