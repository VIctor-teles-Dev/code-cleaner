import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Aplicacoes from "./page";

const params = Promise.resolve({ locale: "pt-BR" });

describe("Aplicacoes", () => {
  it("renders the page heading", async () => {
    render(await Aplicacoes({ params }));

    expect(
      screen.getByRole("heading", { level: 1, name: "Aplicações" }),
    ).toBeInTheDocument();
  });

  it("links online apps to their subdomain", async () => {
    render(await Aplicacoes({ params }));

    expect(
      screen.getByRole("link", { name: /code-cleaner/i }),
    ).toHaveAttribute("href", "http://code-cleaner.ccl.app.br");
  });

  it("shows non-linked apps (building / soon) with their stack", async () => {
    render(await Aplicacoes({ params }));

    expect(screen.getByText("mobile-app")).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /mobile-app/i }),
    ).not.toBeInTheDocument();
    expect(screen.getAllByText(/em breve/i).length).toBeGreaterThan(0);
    expect(screen.getByText(/em construção/i)).toBeInTheDocument();
    expect(screen.getByText("React Native")).toBeInTheDocument();
  });
});
