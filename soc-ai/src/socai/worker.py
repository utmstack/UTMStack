"""Takes new alerts from NATS and triages them.

The backend publishes {"tenantId", "alertId", "facts"} to socai.alerts when an alert is created; facts
is the alert's dossier, already worked out next to the data (see backend/modules/socai/dossier.go). The stream
keeps them while this service is down or busy, and a durable consumer hands them out at the pace the
model allows, so a burst of alerts becomes a queue and not a pile of lost work.
"""

from __future__ import annotations

import asyncio
import json
import logging
from typing import Any

from .limits import is_rate_limited
from .runtime import Runtime

log = logging.getLogger("socai.worker")

STREAM = "SOCAI_ALERTS"
SUBJECT = "socai.alerts"
DURABLE = "soc-ai"
MAX_DELIVER = 3
RETRY_DELAY = 30.0
RATE_LIMIT_DELAY = 120.0
ACK_WAIT = 600.0


async def handle(rt: Runtime, msg: Any) -> None:
    """One message: ack what is done, retry what may work later, drop what never will."""
    try:
        body = json.loads(msg.data)
        alert_id, tenant = body["alertId"], body.get("tenantId")
        facts = body.get("facts")
        facts = facts if isinstance(facts, dict) else None
    except (ValueError, KeyError, TypeError, AttributeError):
        log.error("dropping a message that is not an alert reference: %r", msg.data[:200])
        await msg.term()

        return

    throttled = False

    try:
        outcome = await rt.triage(alert_id, tenant, facts)
    except Exception as exc:  # noqa: BLE001 - usually the model's provider: worth trying again later
        throttled = is_rate_limited(exc)

        if throttled:
            log.warning("triage of %s was rate limited by the model's provider; trying again later", alert_id)
        else:
            log.exception("triage of %s failed", alert_id)

        outcome = "failed: error"

    if not outcome.startswith("failed"):
        await msg.ack()

        return

    delivered = getattr(getattr(msg, "metadata", None), "num_delivered", 1)

    # A throttled model asks for patience, not for the alert to be given up on: wait longer before the next try
    if throttled:
        await msg.nak(delay=RATE_LIMIT_DELAY)
    elif delivered >= MAX_DELIVER:
        log.error("giving up on %s after %d attempts: %s", alert_id, delivered, outcome)
        await msg.term()
    else:
        await msg.nak(delay=RETRY_DELAY)


async def run(rt: Runtime) -> None:
    import nats
    from nats.js.api import ConsumerConfig, RetentionPolicy, StorageType, StreamConfig

    nc = await nats.connect(rt.settings.nats_url, max_reconnect_attempts=-1, name="soc-ai")
    js = nc.jetstream()

    try:
        await js.add_stream(StreamConfig(name=STREAM, subjects=[SUBJECT], retention=RetentionPolicy.LIMITS,
                                         storage=StorageType.FILE, max_age=24 * 3600))
    except Exception as exc:  # noqa: BLE001 - the backend may have created it already, with its own settings
        log.info("stream %s not created here: %s", STREAM, exc)

    sub = await js.pull_subscribe(SUBJECT, durable=DURABLE, stream=STREAM,
                                  config=ConsumerConfig(ack_wait=ACK_WAIT, max_deliver=MAX_DELIVER,
                                                        max_ack_pending=rt.settings.triage_concurrency))
    gate = asyncio.Semaphore(rt.settings.triage_concurrency)
    running: set[asyncio.Task] = set()

    async def one(m: Any) -> None:
        async with gate:
            await handle(rt, m)

    while True:
        try:
            batch = await sub.fetch(rt.settings.triage_concurrency, timeout=5)
        except TimeoutError:
            continue
        except Exception:  # noqa: BLE001
            log.exception("fetching alerts failed; trying again")
            await asyncio.sleep(2)

            continue

        for m in batch:
            t = asyncio.create_task(one(m))
            running.add(t)
            t.add_done_callback(running.discard)
