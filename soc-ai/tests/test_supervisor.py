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
    delegate = route(action="delegate", agent="alerts", task="again")
    model = Scripted(script=[
        delegate, call("count_open"), say("42"),
        delegate, say("nothing new"),  # the repeat of the same call is refused inside the specialist
        delegate, say("nothing new"),
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
