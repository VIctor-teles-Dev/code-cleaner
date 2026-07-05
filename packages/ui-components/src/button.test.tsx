import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Button } from "./button";

describe("Button", () => {
  it("renders its children", () => {
    render(<Button>Salvar</Button>);

    expect(screen.getByRole("button", { name: "Salvar" })).toBeInTheDocument();
  });

  it("defaults to type=button to avoid accidental form submits", () => {
    render(<Button>Salvar</Button>);

    expect(screen.getByRole("button")).toHaveAttribute("type", "button");
  });

  it("exposes the variant for styling", () => {
    render(<Button variant="secondary">Cancelar</Button>);

    expect(screen.getByRole("button")).toHaveAttribute(
      "data-variant",
      "secondary",
    );
  });

  it("uses the primary variant by default", () => {
    render(<Button>Salvar</Button>);

    expect(screen.getByRole("button")).toHaveAttribute(
      "data-variant",
      "primary",
    );
  });
});
