import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { SiteHeader } from "./site-header";

vi.mock("next/navigation", () => ({
  usePathname: () => "/",
}));

describe("SiteHeader", () => {
  it("renders the brand link and main navigation", () => {
    render(<SiteHeader />);

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

  it("marks the current route with aria-current", () => {
    render(<SiteHeader />);

    expect(screen.getByRole("link", { name: "Início" })).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(screen.getByRole("link", { name: "Blog" })).not.toHaveAttribute(
      "aria-current",
    );
  });
});
