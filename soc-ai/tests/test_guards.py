"""What the generic tools refuse or change before anything reaches the backend."""

import json

from fakes import FakeBackend
from socai.generic import check_before, mcp_tools

SPECS = [
    {"name": "soar.rule.create", "description": "d", "read_only": False, "schema": {"type": "object", "properties": {}}},
    {"name": "visualizations.create", "description": "d", "read_only": False, "schema": {"type": "object", "properties": {}}},
]


def widget(dim="dataSource", chart="category", **extra):
    spec = {"dataset": "logs", "chart": chart, "dimension": dim, **extra}

    return {"dashboard_id": "d1", "title": "t", "spec": json.dumps(spec)}


def test_a_rule_is_never_created_switched_on():
    args, refused = check_before("soar.rule.create", {"name": "x", "active": True, "nodes": {}}, {})

    assert refused is None and args["active"] is False


def test_a_correlation_rule_is_never_created_active():
    args, _ = check_before("correlation_rule.create", {"name": "x", "ruleActive": True}, {})

    assert args["ruleActive"] is False


def test_a_mail_node_needs_a_real_address_and_the_model_is_told_to_ask():
    node = {"executor": "mail", "params": {"to": "{{alert.user.email}}", "subject": "s"}}
    _, refused = check_before("soar.rule.create", {"name": "x", "nodes": {"mail": node}}, {})

    assert refused and "Ask the user" in refused

    node["params"]["to"] = "soc@corp.example"
    _, refused = check_before("soar.rule.create", {"name": "x", "nodes": {"mail": node}}, {})
    assert refused is None


async def test_the_same_widget_is_not_created_twice_but_a_different_one_is():
    b = FakeBackend({"visualizations.create": {"id": "w1"}})
    tools = mcp_tools(b, SPECS, ("visualizations.",), tenant="t", allow_changes=True)
    create = tools["visualizations__create"]

    await create.ainvoke(widget())
    again = await create.ainvoke(widget(limit=5, title="other title"))
    assert "already has a widget" in again and b.names().count("visualizations.create") == 1

    await create.ainvoke(widget(dim="dataType"))
    await create.ainvoke(widget(chart="time", dim="", range="7d"))
    await create.ainvoke(widget(chart="time", dim="", range="24h"))
    assert b.names().count("visualizations.create") == 4, "other dimension, other window: different questions"


async def test_a_rule_made_for_real_goes_out_inactive():
    b = FakeBackend({"soar.rule.create": {"relPath": "r.yaml"}})
    tools = mcp_tools(b, SPECS, ("soar.",), tenant="t", allow_changes=True)

    await tools["soar__rule__create"].ainvoke({"name": "x", "active": True, "conditions": [], "roots": [], "nodes": {}})

    assert b.calls[-1][1]["active"] is False
