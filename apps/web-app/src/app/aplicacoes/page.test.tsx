import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Aplicacoes from "./page";

describe("Aplicacoes", () => {
  it("renders the page heading", () => {
    render(<Aplicacoes />);

    expect(
      screen.getByRole("heading", { level: 1, name: "Aplicações" }),
    ).toBeInTheDocument();
  });

  it("links online apps to their subdomain", () => {
    render(<Aplicacoes />);

    expect(
      screen.getByRole("link", { name: /code-cleaner/i }),
    ).toHaveAttribute("href", "http://code-cleaner.ccl.app.br");
  });

  it("shows upcoming apps without links and with their stack", () => {
    render(<Aplicacoes />);

    expect(screen.getByText("mobile-app")).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /mobile-app/i }),
    ).not.toBeInTheDocument();
    expect(screen.getAllByText(/em breve/i).length).toBeGreaterThan(0);
    expect(screen.getByText("React Native")).toBeInTheDocument();
  });
});
