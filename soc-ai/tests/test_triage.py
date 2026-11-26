from socai.agents.registry import build_subagents
from socai.config import Settings
from socai.triage import MARK, Verdict, format_note, triage_graph
from fakes import FakeBackend, Scripted, call, say

ALERT = {"id": "a1", "name": "Impossible travel", "severity": "high", "status": 0, "notes": "",
         "adversary": {"ip": "9.9.9.9", "user": "mallory", "email": "mallory@evil.com"},
         "target": {"ip": "10.0.0.5", "host": "srv", "user": "alice"}}


def backend(notes=""):
    return FakeBackend({
        "store.search": lambda a: {"hits": [{**ALERT, "notes": notes}], "total": 1},
        "alerts.score": {"alertId": "a1", "finalScore": 77, "decision": "COMPLETE"},
        "alerts.update_notes": {"alert_id": "a1"},
        "alerts.update_status": {"updated": 1},
        "alerts.convert_to_incident": {"incident_id": 5},
    })


def verdict(**over):
    base = dict(classification="possible incident", confident=True, adjustment=3,
                threat_assessment="Login from two countries", affected_asset="srv", context="score 77",
                analysis="travel is impossible", action="reset credentials")

    return call("Verdict", **{**base, **over})


def run(model, b, settings=Settings(), alert_id="a1"):
    agents = build_subagents(b, settings, [], groups=settings.triage_groups, hide_people=True)

    return triage_graph(model, b, settings, agents).ainvoke({"alert_id": alert_id})


def test_the_note_is_the_exact_line_the_alerts_page_parses():
    v = Verdict(classification="possible false positive", confident=True, score=12,
                threat_assessment="A | pipe\nand a newline", affected_asset="N/A", context="c", analysis="a", action="close")

    assert format_note(v, 12) == (f"{MARK} Score: 12/100 - Completed - Low Risk | Threat Assessment: A / pipe and a newline "
                              "| Affected Asset: N/A | Context: c | LLM Analysis: a | Action: close")


async def test_a_confident_verdict_is_one_model_call_and_one_note():
    b, model = backend(), Scripted(script=[verdict()])
    out = await run(model, b)

    assert out["outcome"] == "written" and len(model.asked) == 1
    notes = [a for t, a in b.calls if t == "alerts.update_notes"]
    assert len(notes) == 1 and notes[0]["notes"].startswith(MARK + " Score: 80/100 - Open - High Risk")
    assert "alerts.update_status" not in b.names(), "nothing is changed unless an administrator allowed it"


async def test_the_model_never_sees_the_names_of_people():
    b, model = backend(), Scripted(script=[verdict()])
    await run(model, b)
    seen = " ".join(str(m.content) for m in model.asked[0])

    assert "mallory" not in seen and "alice" not in seen and "evil.com" not in seen
    assert "John Doe" in seen and "anonymizedFields" in seen


async def test_an_alert_already_assessed_is_left_alone():
    b, model = backend(notes=f"{MARK} Score: 50/100 - Open - Low Risk | ..."), Scripted(script=[])
    out = await run(model, b)

    assert out["outcome"].startswith("skipped") and model.asked == [] and "alerts.update_notes" not in b.names()


async def test_an_unsure_verdict_investigates_once_on_the_precise_doubt_then_decides():
    b = backend()
    model = Scripted(script=[
        verdict(confident=False, missing="whether 9.9.9.9 attacked other hosts"),
        call("get_alert_dossier", alert_id="a1"),            # the specialist's one lookup
        say("9.9.9.9 only touched this host."),               # its report
        verdict(adjustment=-10, classification="possible false positive"),
    ])
    out = await run(model, b)

    assert out["outcome"] == "written"
    assert "whether 9.9.9.9 attacked other hosts" in str(model.asked[1][-1].content)
    assert "9.9.9.9 only touched this host." in str(model.asked[-1][-1].content), "the final verdict saw the finding"
    assert any("Score: 67/100 - Completed" in a.get("notes", "") for t, a in b.calls if t == "alerts.update_notes")


async def test_it_does_not_investigate_twice():
    b = backend()
    model = Scripted(script=[verdict(confident=False, missing="x"), say("could not tell"), verdict(confident=False)])
    out = await run(model, b)

    assert out["outcome"] == "written" and len(model.asked) == 3


async def test_a_firm_false_positive_is_closed_only_when_allowed():
    off = backend()
    await run(Scripted(script=[verdict(classification="possible false positive")]), off)
    assert "alerts.update_status" not in off.names()

    on = backend()
    await run(Scripted(script=[verdict(classification="possible false positive")]), on,
              Settings(triage_groups=frozenset({"alert_actions"})))
    assert "alerts.update_status" in on.names()


async def test_an_incident_is_opened_only_when_allowed_and_the_verdict_is_firm():
    on = backend()
    await run(Scripted(script=[verdict()]), on, Settings(triage_groups=frozenset({"incident_actions"})))
    assert "alerts.convert_to_incident" in on.names()


async def test_an_alert_that_does_not_exist_ends_cleanly():
    b = FakeBackend({"store.search": {"hits": [], "total": 0}, "alerts.score": {"score": 1}})
    out = await run(Scripted(script=[]), b, alert_id="nope")

    assert out["outcome"].startswith("failed")


async def test_a_dossier_sent_with_the_alert_saves_every_lookup():
    b = FakeBackend({"alerts.update_notes": {"alert_id": "a1"}})
    model = Scripted(script=[verdict()])
    facts = {"alert": {"id": "a1", "name": "Impossible travel", "notes": "",
                       "adversary": {"ip": "9.9.9.9", "user": "mallory"}},
             "score": {"score": 77}, "same_name_30d": 3}
    agents = build_subagents(b, Settings(), [], groups=frozenset(), hide_people=True)
    out = await triage_graph(model, b, Settings(), agents).ainvoke({"alert_id": "a1", "facts": facts})

    assert out["outcome"] == "written"
    assert b.names() == ["alerts.update_notes"], "nothing is asked of the backend but writing the note"
    seen = " ".join(str(m.content) for m in model.asked[0])
    assert "mallory" not in seen and "John Doe" in seen and "same_name_30d" in seen


async def test_a_dossier_that_shows_the_alert_already_assessed_is_respected():
    b = FakeBackend()
    facts = {"alert": {"id": "a1", "notes": f"analyst text\n{MARK} Score: 5/100 - Open - Low Risk | ..."}}
    agents = build_subagents(b, Settings(), [], groups=frozenset(), hide_people=True)
    out = await triage_graph(Scripted(script=[]), b, Settings(), agents).ainvoke({"alert_id": "a1", "facts": facts})

    assert out["outcome"].startswith("skipped") and b.calls == []


async def test_an_alert_a_person_asked_about_is_assessed_again_even_if_it_was_before():
    b = backend(notes=f"{MARK} Score: 50/100 - Open - Low Risk | ...")
    model = Scripted(script=[verdict()])
    agents = build_subagents(b, Settings(), [], groups=frozenset(), hide_people=True)
    out = await triage_graph(model, b, Settings(), agents).ainvoke({"alert_id": "a1", "force": True})

    assert out["outcome"] == "written" and len(model.asked) == 1


def test_the_score_is_the_backends_moved_by_at_most_ten_and_the_risk_follows_it():
    from socai.triage import final_score, risk_for

    facts = {"score": {"finalScore": 33, "decision": "COMPLETE"}}
    v = lambda adj: Verdict(classification="standard alert", confident=True, adjustment=adj, threat_assessment="t",
                            context="c", analysis="a", action="x")

    assert [final_score(v(a), facts) for a in (-10, 0, 7)] == [23, 33, 40], "the same alert can no longer land anywhere"
    assert final_score(v(10), {"score": {"finalScore": 97}}) == 100 and final_score(v(-10), {"score": {"finalScore": 4}}) == 0
    assert final_score(v(0), {"score": {"score": 61}}) == 61, "the older shape still works"
    assert (risk_for(33), risk_for(40), risk_for(69), risk_for(70)) == ("Low", "Medium", "Medium", "High")

    # with no score to start from, the model's own is used, and 50 if it gave none
    own = Verdict(classification="standard alert", confident=True, score=61, threat_assessment="t", context="c", analysis="a", action="x")
    assert final_score(own, {}) == 61
    assert final_score(v(0), {}) == 50


def test_an_adjustment_beyond_ten_is_not_accepted():
    import pytest
    from pydantic import ValidationError

    with pytest.raises(ValidationError):
        Verdict(classification="standard alert", confident=True, adjustment=40, threat_assessment="t", context="c",
                analysis="a", action="x")
