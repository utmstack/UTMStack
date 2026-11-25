from socai.anonymize import hide_people
from socai.shaping import project, render, shape


def test_empties_are_dropped_and_long_lists_say_what_was_left_out():
    out = shape({"a": None, "b": "", "c": [], "d": {"e": None}, "rows": list(range(30))}, max_items=5)

    assert "a" not in out and "b" not in out and "c" not in out and "d" not in out
    assert out["rows"][:5] == [0, 1, 2, 3, 4]
    assert "25 more" in out["rows"][5]["_note"]


def test_long_text_is_clipped():
    assert len(shape("x" * 5000, max_string=100)) < 140


def test_nothing_longer_than_the_budget_reaches_the_model():
    assert len(render({"k": ["y" * 300] * 100}, max_chars=500)) < 600


def test_project_keeps_the_nesting():
    doc = {"id": "1", "adversary": {"ip": "1.2.3.4", "user": "bob", "x": 1}, "junk": 1}

    assert project(doc, ["id", "adversary.ip", "target.ip"]) == {"id": "1", "adversary": {"ip": "1.2.3.4"}}


def test_people_are_hidden_and_listed():
    out, hidden = hide_people({"adversary": {"user": "alice", "ip": "1.2.3.4"}, "message": "mail bob@corp.com now"})

    assert out["adversary"] == {"user": "John Doe", "ip": "1.2.3.4"}
    assert "bob@corp.com" not in out["message"]
    assert set(hidden) == {"adversary.user", "message"}
