import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Encurtador from "./page";

const params = Promise.resolve({ locale: "pt-BR" });

describe("Encurtador", () => {
  it("renders the page heading", async () => {
    render(await Encurtador({ params }));

    expect(
      screen.getByRole("heading", { level: 1, name: /encurtador de url/i }),
    ).toBeInTheDocument();
  });

  it("shows the create form fields", async () => {
    render(await Encurtador({ params }));

    expect(screen.getByLabelText(/url de destino/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/alias custom/i)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /encurtar/i }),
    ).toBeInTheDocument();
  });

  it("exposes the analytics lookup", async () => {
    render(await Encurtador({ params }));

    expect(
      screen.getByRole("button", { name: /ver métricas/i }),
    ).toBeInTheDocument();
  });
});
