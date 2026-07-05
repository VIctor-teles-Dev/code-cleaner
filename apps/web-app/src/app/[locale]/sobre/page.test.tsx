import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Sobre from "./page";

const params = Promise.resolve({ locale: "pt-BR" });

describe("Sobre", () => {
  it("renders the page heading and bio", async () => {
    render(await Sobre({ params }));

    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(
      /sobre mim/i,
    );
    expect(screen.getByText(/problemas reais/i)).toBeInTheDocument();
  });

  it("renders the technology cards", async () => {
    render(await Sobre({ params }));

    for (const tech of ["Go", "TypeScript", "PostgreSQL", "Kubernetes"]) {
      expect(screen.getByText(tech)).toBeInTheDocument();
    }
  });
});
