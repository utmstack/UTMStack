import { beforeEach, describe, expect, it, vi } from "vitest";

const refreshSession = vi.fn();
const getStoredTokens = vi.fn();

vi.mock("@/shared/lib/api-client", () => ({
  getStoredTokens: () => getStoredTokens(),
  refreshSession: () => refreshSession(),
}));
vi.mock("@/shared/config/mode", () => ({ IS_FEDERATION: false }));
vi.mock("@/shared/lib/current-instance", () => ({ getCurrentInstanceId: () => null }));

import { streamChat, type ChatEvent } from "./chat-stream";

const sse = (...events: object[]) =>
  new Response(events.map((e) => `data: ${JSON.stringify(e)}\n\n`).join(""), { status: 200 });

describe("streamChat", () => {
  beforeEach(() => {
    refreshSession.mockReset();
    getStoredTokens.mockReset().mockReturnValue({ access_token: "old" });
    vi.unstubAllGlobals();
  });

  it("renews an expired token once and retries", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ error: "invalid token" }), { status: 401 }))
      .mockResolvedValueOnce(sse({ kind: "final", text: "hello" }));
    vi.stubGlobal("fetch", fetchMock);
    refreshSession.mockResolvedValue("new");

    const events: ChatEvent[] = [];
    await streamChat({ task: "hi" }, (e) => events.push(e));

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(fetchMock.mock.calls[0][1].headers.Authorization).toBe("Bearer old");
    expect(fetchMock.mock.calls[1][1].headers.Authorization).toBe("Bearer new");
    expect(events).toEqual([{ kind: "final", text: "hello" }]);
  });

  it("reports the error when the session cannot be renewed", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: "invalid token" }), { status: 401 })));
    refreshSession.mockResolvedValue(null);

    await expect(streamChat({ task: "hi" }, () => {})).rejects.toThrow("invalid token");
    expect(refreshSession).toHaveBeenCalledTimes(1);
  });

  it("does not refresh when the token is fine", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(sse({ kind: "final", text: "ok" })));

    await streamChat({ task: "hi" }, () => {});
    expect(refreshSession).not.toHaveBeenCalled();
  });
});
