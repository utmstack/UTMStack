import json
from types import SimpleNamespace

from socai import worker
from socai.config import Settings
from socai.runtime import Runtime
from fakes import FakeBackend, Scripted


class Msg:
    def __init__(self, data, delivered=1):
        self.data = data if isinstance(data, bytes) else json.dumps(data).encode()
        self.metadata = SimpleNamespace(num_delivered=delivered)
        self.log = []

    async def ack(self): self.log.append("ack")
    async def nak(self, delay=None): self.log.append(("nak", delay))
    async def term(self): self.log.append("term")


class Rt(Runtime):
    def __init__(self, outcome):
        super().__init__(Settings(), FakeBackend(), Scripted())
        self.outcome, self.seen = outcome, []

    async def triage(self, alert_id, tenant, facts=None):
        self.seen.append((alert_id, tenant, facts))

        if isinstance(self.outcome, Exception):
            raise self.outcome

        return self.outcome


async def test_a_done_alert_is_acknowledged():
    rt, m = Rt("written"), Msg({"alertId": "a1", "tenantId": "t1"})
    await worker.handle(rt, m)

    assert m.log == ["ack"] and rt.seen == [("a1", "t1", None)]


async def test_an_alert_already_assessed_is_acknowledged_too():
    m = Msg({"alertId": "a1"})
    await worker.handle(Rt("skipped: already assessed"), m)

    assert m.log == ["ack"]


async def test_a_failure_is_retried_later_and_given_up_after_the_last_attempt():
    first, last = Msg({"alertId": "a1"}, delivered=1), Msg({"alertId": "a1"}, delivered=worker.MAX_DELIVER)
    await worker.handle(Rt("failed: provider down"), first)
    await worker.handle(Rt(RuntimeError("boom")), last)

    assert first.log == [("nak", worker.RETRY_DELAY)] and last.log == ["term"]


async def test_garbage_is_dropped_not_retried_forever():
    for data in (b"not json", b'{"nope": 1}', b"[]"):
        m = Msg(data)
        await worker.handle(Rt("written"), m)

        assert m.log == ["term"]


async def test_the_dossier_the_backend_sent_goes_on_to_the_triage():
    rt, m = Rt("written"), Msg({"alertId": "a1", "tenantId": "t1", "facts": {"alert": {"id": "a1"}}})
    await worker.handle(rt, m)

    assert rt.seen == [("a1", "t1", {"alert": {"id": "a1"}})]


async def test_facts_that_are_not_an_object_are_ignored_not_fatal():
    rt, m = Rt("written"), Msg({"alertId": "a1", "facts": "garbage"})
    await worker.handle(rt, m)

    assert m.log == ["ack"] and rt.seen[0][2] is None
