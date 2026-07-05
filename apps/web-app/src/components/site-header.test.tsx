import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { SiteHeader } from "./site-header";

describe("SiteHeader", () => {
  it("renders the brand link and main navigation", async () => {
    render(await SiteHeader());

    expect(
      screen.getByRole("link", { name: /página inicial/i }),
    ).toHaveAttribute("href", "/");

    const nav = screen.getByRole("navigation", {
      name: /navegação principal/i,
    });
    expect(nav).toBeInTheDocument();

    for (const label of ["Início", "Sobre", "Aplicações", "Blog", "Contato"]) {
      expect(screen.getByRole("link", { name: label })).toBeInTheDocument();
    }
  });

  it("marks the current route with aria-current", async () => {
    render(await SiteHeader());

    expect(screen.getByRole("link", { name: "Início" })).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(screen.getByRole("link", { name: "Blog" })).not.toHaveAttribute(
      "aria-current",
    );
  });
});
