import re
from datetime import datetime

import pytest

from socai.backend import ToolError
from socai.tools import Toolbox, action_tools, read_tools, since_filter
from fakes import FakeBackend

ALERT = {"id": "a1", "name": "Brute force", "severity": "high", "status": 0, "tenantId": "t",
         "adversary": {"ip": "9.9.9.9", "user": "mallory"}, "target": {"ip": "10.0.0.5", "host": "srv"},
         "lastEvent": {"huge": "x" * 10000}, "notes": ""}


def backend(**over):
    def search(args):
        fields = {f["field"] for f in args["filters"]}

        if "id" in fields:
            return {"hits": [ALERT], "total": 1}

        if "name" in fields:  # how many alerts carry this name: only the total matters
            return {"hits": [], "total": 7}

        return {"hits": [{"id": "a2", "name": "Scan", "adversary": {"ip": "9.9.9.9"}}, ALERT], "total": 2}

    return FakeBackend({"store.search": search,
                        "alerts.score": {"score": 81, "phases": {"intrinsic": 40}}, **over})


async def test_one_call_returns_the_alert_its_score_and_what_surrounds_it():
    b = backend()
    out = await Toolbox(b).alert_dossier("a1")

    assert out["alert"]["name"] == "Brute force"
    assert "lastEvent" not in out["alert"], "the raw event must not be handed over"
    assert out["score"]["score"] == 81
    assert [r["id"] for r in out["same_adversary"]] == ["a2"], "the alert itself is not its own neighbour"
    assert out["same_name_30d"] == 7


async def test_an_unknown_alert_is_an_error_not_an_empty_dossier():
    b = FakeBackend({"store.search": {"hits": [], "total": 0}, "alerts.score": ToolError("nope")})

    with pytest.raises(ToolError):
        await Toolbox(b).alert_dossier("zzz")


async def test_a_failing_side_lookup_does_not_lose_the_dossier():
    b = backend(**{"alerts.score": ToolError("scoring is down")})

    assert "score" not in await Toolbox(b).alert_dossier("a1")


async def test_find_alerts_builds_the_filters_for_the_model():
    b = backend()
    await Toolbox(b).find_alerts(status="open", severity="high", since="7d", limit=100)
    args = b.calls[-1][1]
    fields = {f["field"]: f for f in args["filters"]}

    assert fields["status"]["value"] == "Open" and fields["severity"]["value"] == "high"
    lo, hi = fields["@timestamp"]["value"]
    assert 6.99 * 86400 < (datetime.fromisoformat(hi.rstrip("Z")) - datetime.fromisoformat(lo.rstrip("Z"))).total_seconds() < 7.01 * 86400
    assert args["size"] == 25, "a model may not ask for a hundred rows"


def _span(since):
    lo, hi = since_filter(since)[0]["value"]

    return (datetime.fromisoformat(hi.rstrip("Z")) - datetime.fromisoformat(lo.rstrip("Z"))).total_seconds(), lo


def test_windows_go_out_as_absolute_times():
    # the backend drops a bound it cannot read, and the query then has no time limit
    for since in ("30m", "24h", "7d", "2w", "today", "2026-11-25"):
        for bound in since_filter(since)[0]["value"]:
            assert re.match(r"^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ$", bound), bound


def test_the_forms_a_model_tends_to_write_are_understood():
    assert abs(_span("16h34m")[0] - (16 * 3600 + 34 * 60)) < 2
    assert abs(_span("1h30m")[0] - 5400) < 2
    assert _span("today")[1].endswith("T00:00:00Z")
    assert _span("2026-11-25")[1] == "2026-11-25T00:00:00Z"
    assert _span("2026-11-25 08:30:00")[1] == "2026-11-25T08:30:00Z"


def test_a_window_the_model_made_up_is_refused():
    with pytest.raises(ToolError):
        since_filter("yesterday")


async def test_changes_need_to_be_allowed():
    b = backend(**{"alerts.update_status": {"updated": 1}})

    with pytest.raises(ToolError):
        await Toolbox(b).set_alert_status(["a1"], "closed")

    assert "alerts.update_status" not in b.names()

    await Toolbox(b, groups=frozenset({"alert_actions"})).set_alert_status(["a1"], "closed", "fp")

    assert b.calls[-1] == ("alerts.update_status", {"alert_ids": ["a1"], "status": 2, "status_observation": "fp"})


async def test_an_assessment_goes_as_it_is_because_the_backend_keeps_the_analysts_text():
    b = backend(**{"alerts.update_notes": {"alert_id": "a1"}})
    await Toolbox(b).add_alert_note("a1", "[AI SOC Agent] Score: 5/100")

    assert b.names() == ["alerts.update_notes"] and b.calls[0][1]["notes"] == "[AI SOC Agent] Score: 5/100"


async def test_a_plain_note_keeps_what_was_already_written():
    b = backend(**{"alerts.update_notes": {"alert_id": "a1"}})
    b.handlers["store.search"] = lambda a: {"hits": [{**ALERT, "notes": "analyst: looking"}], "total": 1}
    await Toolbox(b).add_alert_note("a1", "looks like a scan")

    assert b.calls[-1][1]["notes"] == "analyst: looking\nlooks like a scan"


async def test_the_model_sees_errors_as_text_instead_of_a_crash():
    b = FakeBackend({"store.search": {"hits": [], "total": 0}, "alerts.score": ToolError("x")})
    tools = read_tools(Toolbox(b))

    assert (await tools["get_alert_dossier"].ainvoke({"alert_id": "zzz"})).startswith("ERROR:")


async def test_with_people_hidden_the_model_never_gets_a_name():
    b = backend()
    text = await read_tools(Toolbox(b, hide_people=True))["get_alert_dossier"].ainvoke({"alert_id": "a1"})

    assert "mallory" not in text and "John Doe" in text and "anonymizedFields" in text


def test_the_action_tools_are_marked():
    assert all(t.description.startswith("ACTION") for t in action_tools(Toolbox(FakeBackend())).values())


async def test_every_call_tells_the_backend_which_groups_to_turn_on():
    b = backend(**{"alerts.update_status": {"updated": 1}})
    groups = frozenset({"alert_actions"})
    await Toolbox(b, groups=groups).alert_dossier("a1")
    await Toolbox(b, groups=groups).set_alert_status(["a1"], "closed")

    assert set(b.groups_seen) == {groups}


async def test_find_alerts_has_no_implicit_window_and_says_which_one_it_used():
    b = backend()
    out = await Toolbox(b).find_alerts(status="open")

    assert not any(f["field"] == "@timestamp" for f in b.calls[-1][1]["filters"]), "an unasked-for window changes the count"
    assert out["window"] == "all time"

    out = await Toolbox(b).find_alerts(status="open", since="24h")
    assert out["window"] == "24h" and any(f["field"] == "@timestamp" for f in b.calls[-1][1]["filters"])


async def test_log_text_search_looks_in_the_raw_line_not_a_column_that_does_not_exist():
    b = FakeBackend({"store.search": {"hits": [], "total": 0}})
    await Toolbox(b).search_logs(text="failed password")

    fields = [f["field"] for f in b.calls[-1][1]["filters"]]
    assert "raw" in fields and "message" not in fields


async def test_a_ranking_of_people_is_anonymised_when_the_toolbox_hides_people():
    rows = [{"value": "Maria Lopez", "count": 9}, {"value": "WIN-1$", "count": 5}, {"value": "root", "count": 3}, {"value": "jdoe@corp.com", "count": 2}]
    b = FakeBackend({"store.property_values": rows})

    shown = await Toolbox(b, hide_people=True).top_values("logs", "origin.user")
    assert [r["value"] for r in shown] == ["User A", "WIN-1$", "root", "User B"]
    assert [r["count"] for r in shown] == [9, 5, 3, 2], "the counts are what matters and they stay"

    assert await Toolbox(b, hide_people=True).top_values("logs", "origin.ip") == rows, "only people fields are touched"
    assert await Toolbox(b).top_values("logs", "origin.user") == rows, "the chat, which is a person's own tool, is not changed"
