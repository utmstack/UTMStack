"""Telling a model that is throttling us apart from one that is broken."""

from __future__ import annotations

RATE_LIMITED = "The AI model is receiving too many requests right now. Please try again in a minute."


def is_rate_limited(exc: BaseException | None) -> bool:
    """True when the error, or anything it was raised from, is a 429 from the model's provider."""
    seen: set[int] = set()

    while exc is not None and id(exc) not in seen:
        seen.add(id(exc))

        if getattr(exc, "status_code", None) == 429 or type(exc).__name__ == "RateLimitError" \
                or type(exc).__name__ == "OpenAIRateLimitError":
            return True

        response = getattr(exc, "response", None)

        if getattr(response, "status_code", None) == 429:
            return True

        exc = exc.__cause__ or exc.__context__

    return False
