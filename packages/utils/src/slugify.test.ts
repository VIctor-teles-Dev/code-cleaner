import { describe, expect, it } from "vitest";

import { slugify } from "./slugify";

describe("slugify", () => {
  it("lowercases and replaces spaces with hyphens", () => {
    expect(slugify("Write Better Codes")).toBe("write-better-codes");
  });

  it("removes accents", () => {
    expect(slugify("Escreva Códigos Melhores")).toBe("escreva-codigos-melhores");
  });

  it("strips characters that are not alphanumeric", () => {
    expect(slugify("TDD: na prática!")).toBe("tdd-na-pratica");
  });

  it("collapses repeated separators and trims edges", () => {
    expect(slugify("  --hello   world--  ")).toBe("hello-world");
  });
});
