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

  it("shows only the url-shortener app, linked to the encurtador UI", () => {
    render(<Aplicacoes />);

    expect(
      screen.getByRole("link", { name: /url-shortener/i }),
    ).toHaveAttribute("href", "/encurtador");

    // demais cards foram removidos
    expect(screen.queryByText("code-cleaner")).not.toBeInTheDocument();
    expect(screen.queryByText("mobile-app")).not.toBeInTheDocument();
    expect(screen.queryByText("docs")).not.toBeInTheDocument();
  });

  it("marks the app online, with no upcoming apps", () => {
    render(<Aplicacoes />);

    expect(screen.getByText("online")).toBeInTheDocument();
    expect(screen.queryByText(/em breve/i)).not.toBeInTheDocument();
  });
});
