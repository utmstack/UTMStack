from langchain_core.messages import HumanMessage
from langchain_core.tools import StructuredTool

from socai.agents.registry import SubAgent
from socai.config import Settings
from socai.supervisor import supervisor
from fakes import Scripted, call, say

ROUTE = "Route"


def route(**args):
    return call(ROUTE, **args)


def agents():
    async def count() -> str:
        return '{"total": 42}'

    tool = StructuredTool.from_function(coroutine=count, name="count_open", description="d")

    return {"alerts": SubAgent("alerts", "alerts specialist", "you are alerts", {"count_open": tool}),
            "incidents": SubAgent("incidents", "incidents specialist", "you are incidents", {})}


def run(model, text="hi"):
    g = supervisor(model, agents(), Settings(), max_hops=3)

    return g.ainvoke({"messages": [HumanMessage(text)], "context": "", "lang": None})


async def test_it_answers_directly_when_it_can():
    model = Scripted(script=[route(action="finish", answer="Hello!")])
    out = await run(model)

    assert out["answer"] == "Hello!" and len(model.asked) == 1, "no specialist should have been woken"


async def test_it_hands_a_task_to_the_right_specialist_and_answers_from_what_came_back():
    model = Scripted(script=[
        route(action="delegate", agent="alerts", task="how many alerts are open?"),
        call("count_open"),                      # the alerts specialist uses its tool
        say("There are 42 open alerts."),         # ...and reports
        lambda msgs: route(action="finish", answer="42 alerts are open."),
    ])
    out = await run(model, "how many open alerts?")

    assert out["answer"] == "42 alerts are open."
    assert "[alerts] There are 42 open alerts." in str(model.asked[-1][-1].content), "the supervisor saw the findings"


async def test_a_specialist_that_does_not_exist_is_not_run_and_the_supervisor_recovers():
    model = Scripted(script=[
        route(action="delegate", agent="astrologer", task="read the stars"),
        route(action="finish", answer="I cannot do that."),
    ])
    out = await run(model)

    assert out["answer"] == "I cannot do that."
    assert "no specialist named 'astrologer'" in str(model.asked[-1][-1].content)


async def test_it_stops_delegating_after_the_hops_it_is_allowed():
    model = Scripted(script=[
        route(action="delegate", agent="alerts", task="count the open alerts"), call("count_open"), say("42"),
        route(action="delegate", agent="alerts", task="break the alerts down by severity"), say("nothing new"),
        route(action="delegate", agent="alerts", task="find the newest critical detection"), say("nothing new"),
        route(action="finish", answer="enough"),
    ])
    out = await run(model)

    assert out["answer"] == "enough"
    assert "used all your delegations" in str(model.asked[-1][-2].content) or "used all your delegations" in str(model.asked[-1][-1].content)


async def test_finishing_without_an_answer_writes_one_from_the_findings():
    model = Scripted(script=[
        route(action="delegate", agent="alerts", task="count"), call("count_open"), say("42 open"),
        route(action="finish"),
        say("There are 42 open alerts."),
    ])
    out = await run(model)

    assert out["answer"] == "There are 42 open alerts."


async def test_the_same_question_in_other_words_is_not_asked_again():
    model = Scripted(script=[
        route(action="delegate", agent="alerts", task="how many open alerts are there right now?"), call("count_open"), say("42"),
        route(action="delegate", agent="alerts", task="how many alerts are open right now"),   # a rewording
        route(action="finish", answer="42 are open."),
    ])
    out = await run(model, "open alerts?")

    assert out["answer"] == "42 are open."
    assert "you already asked alerts this" in str(model.asked[-1][-1].content)
    assert len(model.asked) == 5, "the second delegation never woke the specialist"


async def test_a_different_task_for_the_same_specialist_still_goes_through():
    model = Scripted(script=[
        route(action="delegate", agent="alerts", task="count the open alerts"), call("count_open"), say("42"),
        route(action="delegate", agent="alerts", task="list the five newest high severity detections"), say("none"),
        route(action="finish", answer="done"),
    ])
    out = await run(model)

    assert out["answer"] == "done" and "[alerts] none" in str(model.asked[-1][-1].content)
