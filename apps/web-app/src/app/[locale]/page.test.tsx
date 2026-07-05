import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Home from "./page";

const params = Promise.resolve({ locale: "pt-BR" });

describe("Home", () => {
  it("renders the hero heading", async () => {
    render(await Home({ params }));

    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(
      /code-cleaner/i,
    );
  });

  it("renders the call-to-action links", async () => {
    render(await Home({ params }));

    expect(
      screen.getByRole("link", { name: /ver projetos/i }),
    ).toHaveAttribute("href", "/aplicacoes");
    expect(screen.getByRole("link", { name: /ler o blog/i })).toHaveAttribute(
      "href",
      "/blog",
    );
  });
});
