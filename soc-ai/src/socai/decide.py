"""Makes a model choose from a fixed shape, so a decision is data and not prose to be parsed."""

from __future__ import annotations

import json
import re
from typing import Any, TypeVar

from langchain_core.language_models.chat_models import BaseChatModel
from pydantic import BaseModel

T = TypeVar("T", bound=BaseModel)


async def decide(model: BaseChatModel, schema: type[T], messages: list[Any]) -> T:
    """Asks for one call of the tool `schema`, which is how every provider returns structured output."""
    bound = model.bind_tools([schema], tool_choice=schema.__name__)
    msg = await bound.ainvoke(messages)

    for call in getattr(msg, "tool_calls", None) or []:
        if call.get("name") == schema.__name__:
            return schema(**call["args"])

    # A model that ignored the tool and answered in text: take the JSON object out of it if there is one.
    text = msg.content if isinstance(msg.content, str) else json.dumps(msg.content)
    m = re.search(r"\{.*\}", text, re.S)

    if m:
        return schema(**json.loads(m.group(0)))

    raise ValueError(f"the model did not return a {schema.__name__}")
