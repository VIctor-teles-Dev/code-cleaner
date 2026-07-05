import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Home from "./page";

describe("Home", () => {
  it("renders the hero heading", () => {
    render(<Home />);

    expect(
      screen.getByRole("heading", { level: 1 }),
    ).toHaveTextContent(/code-cleaner/i);
  });

  it("renders the call-to-action links", () => {
    render(<Home />);

    expect(
      screen.getByRole("link", { name: /ver projetos/i }),
    ).toHaveAttribute("href", "/aplicacoes");
    expect(screen.getByRole("link", { name: /ler o blog/i })).toHaveAttribute(
      "href",
      "/blog",
    );
  });
});
