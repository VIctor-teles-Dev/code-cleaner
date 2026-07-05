import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Sobre from "./page";

describe("Sobre", () => {
  it("renders the page heading and bio", () => {
    render(<Sobre />);

    expect(
      screen.getByRole("heading", { level: 1 }),
    ).toHaveTextContent(/sobre mim/i);
    expect(screen.getByText(/problemas reais/i)).toBeInTheDocument();
  });

  it("renders the technology cards", () => {
    render(<Sobre />);

    const list = screen.getByRole("list");
    expect(list).toBeInTheDocument();

    for (const tech of ["Go", "TypeScript", "PostgreSQL", "Kubernetes"]) {
      expect(screen.getByText(tech)).toBeInTheDocument();
    }
  });
});
