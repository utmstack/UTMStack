import asyncio
import json
from types import SimpleNamespace

import pytest

from socai import config_source
from socai.config_source import ConfigSource
from socai.crypto import Cipher
from socai.model_config import ModelConfig

C = Cipher("the-key")


def stored(model="gpt"):
    return C.encrypt(json.dumps({"default": {"provider": "openai", "model": model, "url": "u"}}))


def test_a_published_configuration_becomes_the_model_of_every_tenant():
    s = ConfigSource(C)
    s.apply(stored())

    assert s.get("t1").model == "gpt"


def test_a_configuration_encrypted_with_another_key_is_refused_and_the_old_one_stays():
    s = ConfigSource(C)
    s.apply(stored("good"))

    with pytest.raises(ValueError):
        s.apply(Cipher("another-key").encrypt(json.dumps({"default": {"provider": "openai", "model": "x", "url": "u"}})))

    assert s.get("t").model == "good"


def test_a_removed_key_means_nothing_is_configured():
    s = ConfigSource(C)
    s.apply(stored())
    s.apply(None)

    assert s.get("t") is None


def test_the_environment_model_is_only_a_fallback():
    fallback = ModelConfig(provider="openai", model="dev-model", url="u")
    s = ConfigSource(C, fallback)

    assert s.get("t").model == "dev-model"

    s.apply(stored("published"))
    assert s.get("t").model == "published"

    s.apply(None)
    assert s.get("t").model == "dev-model"


class Entry(SimpleNamespace):
    pass


class FakeWatcher:
    def __init__(self, script):
        self.script = script  # shared: the test adds what the server "publishes" later

    async def updates(self, timeout=None):
        while not self.script:
            await asyncio.sleep(0.01)

        item = self.script.pop(0)

        if isinstance(item, Exception):
            raise item

        return item


class FakeKV:
    def __init__(self, script):
        self.script = script

    async def watch(self, key):
        assert key == config_source.KEY

        return FakeWatcher(self.script)


class FakeJS:
    def __init__(self, kv, missing=0):
        self.kv, self.missing, self.asked = kv, missing, 0

    async def key_value(self, bucket):
        assert bucket == config_source.BUCKET
        self.asked += 1

        if self.asked <= self.missing:
            raise RuntimeError("bucket not found")

        return self.kv


def fake_connect(js):
    async def connect(url, **kw):
        return SimpleNamespace(jetstream=lambda: js)

    return connect


async def run_watch(s, js):
    task = asyncio.create_task(s.watch("nats://x", connect=fake_connect(js)))
    assert await s.wait_until_read(5), "the configuration was never read"

    return task


async def test_the_watch_follows_the_key_through_changes_and_removal(monkeypatch):
    monkeypatch.setattr(config_source, "RETRY", 0.01)
    s = ConfigSource(C)
    js = FakeJS(FakeKV([Entry(operation="PUT", value=stored("one").encode()), None]))
    task = await run_watch(s, js)

    assert s.get("t").model == "one"

    js.kv.script.append(Entry(operation="PUT", value=stored("two").encode()))
    await asyncio.sleep(0.05)
    assert s.get("t").model == "two"

    js.kv.script.append(Entry(operation="PURGE", value=b""))
    await asyncio.sleep(0.05)
    assert s.get("t") is None

    task.cancel()


async def test_a_value_that_cannot_be_read_does_not_stop_the_watch_or_replace_the_good_one(monkeypatch):
    monkeypatch.setattr(config_source, "RETRY", 0.01)
    s = ConfigSource(C)
    js = FakeJS(FakeKV([Entry(operation="PUT", value=stored("good").encode()), None,
                        Entry(operation="PUT", value=b"garbage that is not a configuration"),
                        Entry(operation="PUT", value=stored("better").encode())]))
    task = await run_watch(s, js)
    await asyncio.sleep(0.1)

    assert s.get("t").model == "better"
    task.cancel()


async def test_it_waits_for_the_bucket_to_exist(monkeypatch):
    monkeypatch.setattr(config_source, "RETRY", 0.01)
    s = ConfigSource(C)
    js = FakeJS(FakeKV([Entry(operation="PUT", value=stored("late").encode()), None]), missing=3)
    task = await run_watch(s, js)

    assert js.asked == 4 and s.get("t").model == "late"
    task.cancel()


async def test_with_no_key_published_yet_it_still_reports_what_it_found(monkeypatch):
    monkeypatch.setattr(config_source, "RETRY", 0.01)
    s = ConfigSource(C)
    task = await run_watch(s, FakeJS(FakeKV([None])))

    assert s.get("t") is None
    task.cancel()


async def test_a_quiet_minute_does_not_restart_the_watch(monkeypatch):
    import nats.errors

    monkeypatch.setattr(config_source, "RETRY", 0.01)
    s = ConfigSource(C)
    # An idle watch raises a timeout from updates(); it must keep the same watcher, not open another.
    js = FakeJS(FakeKV([Entry(operation="PUT", value=stored("one").encode()), None, nats.errors.TimeoutError(),
                        nats.errors.TimeoutError()]))
    task = await run_watch(s, js)
    js.kv.script.append(Entry(operation="PUT", value=stored("two").encode()))
    await asyncio.sleep(0.1)

    assert s.get("t").model == "two" and js.asked == 1, "the key_value bucket was asked for once"
    task.cancel()
