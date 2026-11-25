"""Where the model configuration comes from: the NATS bucket the backend keeps it in.

The service watches the key, so it has the current configuration when it starts and every change after
it, with no restart. Until the backend has published anything there is no model, and work that needs
one waits (an alert is retried later) or is refused with a clear message (the chat).
"""

from __future__ import annotations

import asyncio
import logging
from typing import Any, Callable

from .crypto import Cipher
from .model_config import ModelConfig, ModelConfigs

log = logging.getLogger("socai.config")

BUCKET = "soc-ai-config"
KEY = "config"
RETRY = 5.0


class ConfigSource:
    def __init__(self, cipher: Cipher | None = None, fallback: ModelConfig | None = None) -> None:
        self._cipher = cipher
        self._fallback = fallback
        self._configs = ModelConfigs()
        self._ready = asyncio.Event()

    def get(self, tenant: str | None) -> ModelConfig | None:
        return self._configs.for_tenant(tenant) or (self._fallback if self._fallback and self._fallback.active else None)

    def apply(self, stored: str | None) -> None:
        """Takes a value from the bucket; None means the key is gone and nothing is configured."""
        if stored is None:
            self._configs = ModelConfigs()
            return

        plain = self._cipher.decrypt(stored) if self._cipher else stored

        if not plain.lstrip().startswith("{"):
            # The cipher hands back what it cannot read: a different key, or something else wrote this.
            raise ValueError("the configuration did not decrypt: is ENCRYPTION_KEY the backend's?")

        self._configs = ModelConfigs.parse(plain, self._cipher)

    async def wait_until_read(self, timeout: float) -> bool:
        try:
            await asyncio.wait_for(self._ready.wait(), timeout)
            return True
        except asyncio.TimeoutError:
            return False

    async def watch(self, nats_url: str, *, connect: Callable[..., Any] | None = None) -> None:
        """Follows the key for as long as the process lives."""
        import nats

        nc = await (connect or nats.connect)(nats_url, max_reconnect_attempts=-1, name="soc-ai-config")
        js = nc.jetstream()

        while True:
            try:
                kv = await js.key_value(BUCKET)
                w = await kv.watch(KEY)

                while True:
                    try:
                        entry = await w.updates(timeout=60)
                    except asyncio.TimeoutError:
                        continue  # nothing changed for a minute; the same watch goes on

                    if entry is None:
                        # The end of what was stored when the watch began.
                        self._ready.set()
                        continue

                    try:
                        self.apply(None if entry.operation in ("DEL", "PURGE") else entry.value.decode())
                    except Exception:  # noqa: BLE001 - keep the configuration that was working
                        log.exception("the model configuration could not be read; keeping the previous one")
            except Exception as exc:  # noqa: BLE001 - NATS or the bucket is not there yet
                log.warning("waiting for the model configuration (%s): %s", BUCKET, exc)
                await asyncio.sleep(RETRY)
