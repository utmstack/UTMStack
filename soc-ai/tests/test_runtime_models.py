import json

import httpx
import pytest

from socai import runtime as runtime_module
from socai.api import create_app
from socai.config import Settings
from socai.config_source import ConfigSource
from socai.crypto import Cipher
from socai.runtime import NotConfigured, Runtime
from fakes import FakeBackend, Scripted

C = Cipher("k")


def configs(**doc):
    s = ConfigSource(C)
    s.apply(C.encrypt(json.dumps(doc)))

    return s


def rt(cfgs, built=None):
    r = Runtime(Settings(internal_key="k"), FakeBackend(), configs=cfgs)
    r.__dict__["_built"] = built

    return r


@pytest.fixture
def counting(monkeypatch):
    built = []

    def fake(cfg, *, temperature=0.0):
        built.append(cfg.model)

        return Scripted()

    monkeypatch.setattr(runtime_module, "build_model", fake)

    return built


def test_each_tenant_runs_on_its_own_model_or_the_default(counting):
    r = rt(configs(default={"provider": "openai", "model": "gpt", "url": "u"},
                   tenants={"t1": {"provider": "openai", "model": "own", "url": "u"}}))

    r.model_for("t1")
    r.model_for("t2")
    r.model_for("t3")
    r.model_for("t1")

    assert counting == ["own", "gpt"], "one model for each configuration, however many tenants share it"


def test_a_tenant_with_nothing_configured_is_told_so(counting):
    with pytest.raises(NotConfigured):
        rt(ConfigSource(C)).model_for("t1")


def test_the_limits_of_the_tenants_configuration_apply_to_its_work(counting):
    r = rt(configs(default={"provider": "openai", "model": "m", "url": "u", "max_tool_iterations": 4,
                            "allow_incidents": True}))
    _, settings = r.model_for("t")

    assert settings.max_steps == 4 and "incident_actions" in settings.triage_groups
    assert "alert_actions" not in settings.triage_groups, "closing alerts is not something the page offers"


def test_without_the_incident_switch_triage_asks_for_no_group(counting):
    _, settings = rt(configs(default={"provider": "openai", "model": "m", "url": "u"})).model_for("t")

    assert settings.triage_groups == frozenset()


def test_a_fixed_model_is_used_for_everyone():
    m = Scripted()
    r = Runtime(Settings(), FakeBackend(), m)

    assert r.model_for("anyone")[0] is m


async def test_the_chat_of_an_unconfigured_tenant_says_why_instead_of_failing(counting):
    app = create_app(rt(ConfigSource(C)))

    async with httpx.AsyncClient(transport=httpx.ASGITransport(app=app), base_url="http://t") as c:
        r = await c.post("/api/v1/agent/task", json={"task": "hi"}, headers={"X-Internal-Key": "k", "X-Tenant-Id": "t1"})

    assert "event: error" in r.text and "not configured" in r.text
