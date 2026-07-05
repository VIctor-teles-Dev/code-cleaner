import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import Blog from "./page";

function stubPosts(posts: unknown) {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      new Response(JSON.stringify(posts), { status: 200 }),
    ),
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("Blog", () => {
  it("lists published posts with links and tags", async () => {
    stubPosts([
      {
        slug: "primeiro-post",
        title: "Primeiro post",
        excerpt: "Um resumo do post.",
        published_at: "2026-07-01T12:00:00Z",
        tags: [{ slug: "minhas-aplicacoes", name: "minhas aplicações" }],
      },
    ]);

    render(await Blog({}));

    expect(
      screen.getByRole("heading", { level: 1, name: "Blog" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: /primeiro post/i }),
    ).toHaveAttribute("href", "/blog/primeiro-post");
    expect(screen.getByText("Um resumo do post.")).toBeInTheDocument();
    expect(screen.getByText(/1 de julho de 2026/)).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "#minhas aplicações" }),
    ).toHaveAttribute("href", "/blog?tag=minhas-aplicacoes");
  });

  it("filters by tag from the query string", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify([]), { status: 200 }),
    );
    vi.stubGlobal("fetch", fetchMock);

    render(
      await Blog({
        searchParams: Promise.resolve({ tag: "minhas-aplicacoes" }),
      }),
    );

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/posts?tag=minhas-aplicacoes"),
      expect.anything(),
    );
    expect(screen.getByText(/filtrando por/i)).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: /limpar filtro/i }),
    ).toHaveAttribute("href", "/blog");
    expect(screen.getByText(/nenhum post com essa tag/i)).toBeInTheDocument();
  });

  it("shows the empty state when there are no posts", async () => {
    stubPosts([]);

    render(await Blog({}));

    expect(screen.getByText(/nenhum post publicado/i)).toBeInTheDocument();
  });
});
