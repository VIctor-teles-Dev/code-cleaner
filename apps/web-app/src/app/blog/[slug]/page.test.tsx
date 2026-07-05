import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import PostPage from "./page";

const params = Promise.resolve({ slug: "primeiro-post" });

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("PostPage", () => {
  it("renders the post with markdown content", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            slug: "primeiro-post",
            title: "Primeiro post",
            content: "## Seção\n\nParágrafo com `código` inline.",
            published_at: "2026-07-01T12:00:00Z",
            tags: [{ slug: "minhas-aplicacoes", name: "minhas aplicações" }],
          }),
          { status: 200 },
        ),
      ),
    );

    render(await PostPage({ params }));

    expect(
      screen.getByRole("heading", { level: 1, name: "Primeiro post" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { level: 2, name: "Seção" }),
    ).toBeInTheDocument();
    expect(screen.getByText("código")).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: /voltar para o blog/i }),
    ).toHaveAttribute("href", "/blog");
    expect(
      screen.getByRole("link", { name: "#minhas aplicações" }),
    ).toHaveAttribute("href", "/blog?tag=minhas-aplicacoes");
  });

  it("throws notFound for a missing post", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ status: "not_found" }), { status: 404 }),
      ),
    );

    await expect(PostPage({ params })).rejects.toThrow();
  });
});
