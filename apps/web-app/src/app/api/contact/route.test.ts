import { afterEach, describe, expect, it, vi } from "vitest";

import { POST } from "./route";

function jsonRequest(body: string) {
  return new Request("http://localhost/api/contact", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body,
  });
}

const validBody = JSON.stringify({
  name: "Ana",
  email: "ana@example.com",
  message: "Olá!",
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("POST /api/contact", () => {
  it("forwards the message to the backend and relays its response", async () => {
    const upstream = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ status: "received" }), { status: 201 }),
    );
    vi.stubGlobal("fetch", upstream);

    const response = await POST(jsonRequest(validBody));

    expect(response.status).toBe(201);
    await expect(response.json()).resolves.toEqual({ status: "received" });
    expect(upstream).toHaveBeenCalledWith(
      expect.stringContaining("/contact"),
      expect.objectContaining({ method: "POST" }),
    );
  });

  it("rejects malformed json without calling the backend", async () => {
    const upstream = vi.fn();
    vi.stubGlobal("fetch", upstream);

    const response = await POST(jsonRequest("{"));

    expect(response.status).toBe(400);
    expect(upstream).not.toHaveBeenCalled();
  });

  it("returns 502 when the backend is unreachable", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new Error("connection refused")),
    );

    const response = await POST(jsonRequest(validBody));

    expect(response.status).toBe(502);
    await expect(response.json()).resolves.toEqual({ status: "error" });
  });
});
