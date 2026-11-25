"""Starts the service: the HTTP API, and the alert worker when NATS is configured."""

from __future__ import annotations

import asyncio
import logging

import uvicorn

from . import worker
from .api import create_app
from .backend import McpBackend
from .config import Settings
from .config_source import ConfigSource
from .crypto import Cipher
from .model_config import ModelConfig
from .runtime import Runtime


async def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s %(message)s")
    s = Settings.from_env()
    backend = McpBackend(s.backend_url, s.internal_key)
    # The model is the tenant's, from the configuration the backend publishes. A model set in the
    # environment is only a fallback for development, where there is no backend to publish one.
    fallback = (ModelConfig(provider=s.llm_provider, model=s.llm_model, url=s.llm_base_url or "", api_key=s.llm_api_key,
                            max_tokens=s.llm_max_tokens) if s.llm_model else None)
    configs = ConfigSource(Cipher(s.encryption_key) if s.encryption_key else None, fallback)
    rt = Runtime(s, backend, configs=configs)
    server = uvicorn.Server(uvicorn.Config(create_app(rt), host="0.0.0.0", port=s.http_port, log_level="info"))
    tasks = [asyncio.create_task(server.serve())]

    if s.nats_url:
        tasks.append(asyncio.create_task(configs.watch(s.nats_url)))
        tasks.append(asyncio.create_task(worker.run(rt)))
    else:
        logging.getLogger("socai").warning("NATS_URL is not set: new alerts will not be triaged, only the API is up")

    try:
        await asyncio.gather(*tasks)
    finally:
        await backend.close()


if __name__ == "__main__":
    asyncio.run(main())
