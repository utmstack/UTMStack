"""The generic tools (dashboards, SOAR, compliance, data sources, rules) must call the backend with the groups
the request turned on. Without them the backend answers for the core tools only, so every one of these came
back as "the tool is not available" and creating a dashboard or a flow never worked in the chat."""

import json

from fakes import FakeBackend
from socai.agents.registry import build_subagents
from socai.config import Settings
from socai.generic import mcp_tools

SPECS = [
    {"name": "dashboards.create", "description": "Creates a dashboard", "read_only": False,
     "schema": {"type": "object", "properties": {"name": {"type": "string"}}, "required": ["name"]}},
    {"name": "soar.rule.list", "description": "Lists rules", "read_only": True, "schema": {"type": "object", "properties": {}}},
    {"name": "store.dataset.fields", "description": "Fields", "read_only": True,
     "schema": {"type": "object", "properties": {"dataset": {"type": "string"}}}},
]


async def test_a_generic_tool_calls_the_backend_with_its_groups():
    b = FakeBackend({"dashboards.create": {"id": "d1"}})
    tools = mcp_tools(b, SPECS, ("dashboards.",), tenant="t", allow_changes=True, groups=frozenset({"dashboards"}))

    await tools["dashboards__create"].ainvoke({"name": "QA"})

    assert b.groups_seen == [frozenset({"dashboards"})]


async def test_the_specialists_pass_their_groups_down_to_every_tool():
    b = FakeBackend({"dashboards.create": {"id": "d1"}, "soar.rule.list": [], "store.dataset.fields": []})
    groups = frozenset({"dashboards", "soar"})
    agents = build_subagents(b, Settings(), SPECS, groups=groups, tenant="t")

    await agents["dashboards"].tools["dashboards__create"].ainvoke({"name": "QA"})
    await agents["soar"].tools["soar__rule__list"].ainvoke({})

    assert b.groups_seen == [groups, groups], "each call must carry the groups, or the backend hides the tool"
