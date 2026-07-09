import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { AnalyticsView } from "./analytics-view";

const sample = {
  slug: "aB3x9",
  total_clicks: 3,
  time_series: [{ day: "2026-07-01", count: 3 }],
  top_countries: [{ label: "BR", count: 3 }],
  top_referrers: [],
  browsers: [{ label: "Firefox", count: 3 }],
  devices: [{ label: "Desktop", count: 3 }],
};

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("AnalyticsView", () => {
  it("fetches the slug and renders the dashboard", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(new Response(JSON.stringify(sample), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    render(<AnalyticsView slug="aB3x9" />);

    // O dashboard só aparece depois que o fetch resolve.
    expect(await screen.findByText(/cliques em/i)).toBeInTheDocument();
    expect(screen.getByText("Navegadores")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/encurtador/analytics/aB3x9");
  });

  it("shows a not-found message for an unknown slug", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ status: "not_found" }), { status: 404 }),
      ),
    );

    render(<AnalyticsView slug="nope" />);

    expect(
      await screen.findByText(/nenhum link com esse código/i),
    ).toBeInTheDocument();
  });

  it("shows a generic error when the request fails", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("network")));

    render(<AnalyticsView slug="aB3x9" />);

    expect(
      await screen.findByText(/não foi possível carregar/i),
    ).toBeInTheDocument();
  });
});
