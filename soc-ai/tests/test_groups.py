from socai.agents.registry import build_subagents
from socai.config import GROUPS, Settings, normalize_groups
from socai.runtime import Runtime
from fakes import FakeBackend, Scripted


def spec(name, read_only=True):
    return {"name": name, "description": name, "schema": {"type": "object", "properties": {}}, "read_only": read_only}


SPECS = [spec("dashboards.list"), spec("dashboards.create", False), spec("soar.rule.list"),
         spec("soar.rule.create", False), spec("compliance.framework.list"), spec("incident_notes.list"),
         spec("incident_notes.create", False), spec("store.dataset.fields"), spec("iam.user.list"),
         spec("tenants.terminate", False)]


def agents(groups):
    return build_subagents(FakeBackend(), Settings(), SPECS, groups=frozenset(groups))


def test_the_old_names_of_the_settings_page_still_mean_something():
    assert normalize_groups(["alerts", "incidents", "correlation", "soar"]) == {
        "alert_actions", "incident_actions", "detection", "soar"}
    assert normalize_groups(["nonsense", "", "dashboards"]) == {"dashboards"}
    assert normalize_groups(None or []) == frozenset()


def test_with_nothing_turned_on_only_the_investigating_specialists_exist():
    assert set(agents([])) == {"alerts", "incidents", "logs"}


def test_a_specialist_appears_only_when_its_group_is_on():
    names = set(agents({"dashboards", "soar"}))

    assert {"dashboards", "soar"} <= names and "compliance" not in names


def test_the_tools_that_change_things_need_their_group():
    off, on = agents({"soar"}), agents({"soar", "incident_actions"})
    assert "soar__rule__create" in off["soar"].tools, "a tool that is listed is offered; the backend decides what is listed"

    assert "soar__rule__create" in on["soar"].tools and "soar__rule__list" in on["soar"].tools
    assert "incident_notes__create" not in off["incidents"].tools
    assert "incident_notes__create" in on["incidents"].tools
    assert "create_incident" not in off["alerts"].tools and "create_incident" in on["alerts"].tools


def test_what_manages_people_or_tenants_is_never_offered_even_if_the_backend_listed_it():
    for a in agents(GROUPS).values():
        assert not any(n.startswith(("iam", "tenants")) for n in a.tools), a.name


async def test_the_runtime_asks_the_backend_for_the_groups_of_the_work_and_remembers_the_answer():
    b = FakeBackend()
    b.specs = SPECS
    rt = Runtime(Settings(triage_groups=frozenset({"incident_actions"})), b, Scripted())

    await rt.agents(tenant=None, groups=frozenset({"incident_actions"}))
    await rt.agents(tenant=None, groups=frozenset({"incident_actions"}))
    await rt.agents(tenant=None, groups=frozenset())

    assert b.listed == [frozenset({"incident_actions"}), frozenset()], "one listing for each set, not for each use"


async def test_the_chat_gets_every_group_unless_the_request_narrows_it():
    b = FakeBackend()
    b.specs = SPECS
    rt = Runtime(Settings(), b, Scripted())

    await rt.chat_graph(tenant=None)
    await rt.chat_graph(tenant=None, groups=frozenset({"soar"}))

    assert b.listed[0] == GROUPS and b.listed[1] == frozenset({"soar"})
