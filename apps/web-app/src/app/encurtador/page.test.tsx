import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Encurtador from "./page";

describe("Encurtador", () => {
  it("renders the page heading", () => {
    render(<Encurtador />);

    expect(
      screen.getByRole("heading", { level: 1, name: /encurtador de url/i }),
    ).toBeInTheDocument();
  });

  it("shows the create form fields", () => {
    render(<Encurtador />);

    expect(screen.getByLabelText(/url de destino/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/alias custom/i)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /encurtar/i }),
    ).toBeInTheDocument();
  });

  it("exposes the analytics lookup", () => {
    render(<Encurtador />);

    expect(
      screen.getByRole("button", { name: /ver métricas/i }),
    ).toBeInTheDocument();
  });
});
