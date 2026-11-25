from langchain_core.messages import HumanMessage, ToolMessage
from langchain_core.tools import StructuredTool

from socai.graph import REPEAT, last_text, react
from fakes import Scripted, call, say


def counter():
    n = {"calls": 0}

    async def run(q: str = "") -> str:
        n["calls"] += 1

        return f"result for {q}"

    return n, {"lookup": StructuredTool.from_function(coroutine=run, name="lookup", description="d")}


async def test_it_answers_after_the_tools_it_needed():
    n, tools = counter()
    model = Scripted(script=[call("lookup", q="a"), say("done")])
    out = await react(model, tools, "sys").ainvoke({"messages": [HumanMessage("go")]})

    assert last_text(out) == "done" and n["calls"] == 1


async def test_a_repeated_call_is_not_run_again():
    n, tools = counter()
    model = Scripted(script=[call("lookup", q="a"), call("lookup", q="a"), say("done")])
    out = await react(model, tools, "sys").ainvoke({"messages": [HumanMessage("go")]})

    assert n["calls"] == 1
    assert any(isinstance(m, ToolMessage) and m.content == REPEAT for m in out["messages"])


async def test_the_total_of_calls_is_capped_and_the_model_is_told():
    n, tools = counter()
    script = [call("lookup", q=str(i)) for i in range(6)] + [say("done")]
    out = await react(Scripted(script=script), tools, "sys", max_tool_calls=3).ainvoke({"messages": [HumanMessage("go")]})

    assert n["calls"] == 3
    assert any(isinstance(m, ToolMessage) and "No more tool calls" in m.content for m in out["messages"])


async def test_when_the_steps_run_out_the_model_must_answer_without_tools():
    n, tools = counter()
    seen_tools = []

    def last(messages):
        seen_tools.append("no steps left" in str(messages[-1].content))

        return say("best effort")

    model = Scripted(script=[call("lookup", q="1"), call("lookup", q="2"), last])
    out = await react(model, tools, "sys", max_steps=3).ainvoke({"messages": [HumanMessage("go")]})

    assert last_text(out) == "best effort" and seen_tools == [True]


async def test_an_unknown_tool_is_reported_to_the_model():
    _, tools = counter()
    out = await react(Scripted(script=[call("nope"), say("ok")]), tools, "sys").ainvoke({"messages": [HumanMessage("go")]})

    assert any(isinstance(m, ToolMessage) and "no tool named nope" in m.content for m in out["messages"])
