import "@testing-library/jest-dom/vitest";

import { cleanup } from "@testing-library/react";
import { afterEach, vi } from "vitest";

afterEach(cleanup);

// i18n nos testes: resolve as chaves usando as mensagens reais pt-BR, tanto no
// caminho de servidor (getTranslations) quanto no de cliente (useTranslations),
// para as asserções continuarem batendo com as strings em português.
const LOCALE = "pt-BR";

vi.mock("next-intl", async (importOriginal) => {
  const actual = await importOriginal<typeof import("next-intl")>();
  const messages = (await import("./messages/pt-BR.json"))
    .default as Record<string, unknown>;
  const make = (ns?: string) =>
    actual.createTranslator({ locale: LOCALE, messages, namespace: ns });
  return {
    ...actual,
    useTranslations: (ns?: string) => make(ns),
    useLocale: () => LOCALE,
  };
});

vi.mock("next-intl/server", async () => {
  const actual = await vi.importActual<typeof import("next-intl")>("next-intl");
  const messages = (await import("./messages/pt-BR.json"))
    .default as Record<string, unknown>;
  const make = (ns?: string) =>
    actual.createTranslator({ locale: LOCALE, messages, namespace: ns });
  return {
    getTranslations: async (arg?: string | { namespace?: string }) =>
      make(typeof arg === "string" ? arg : arg?.namespace),
    setRequestLocale: () => {},
    getLocale: async () => LOCALE,
  };
});

// Navegação locale-aware vira <a> simples nos testes.
vi.mock("@/i18n/navigation", async () => {
  const React = await import("react");
  return {
    Link: ({
      href,
      children,
      ...props
    }: {
      href: unknown;
      children?: unknown;
    }) =>
      React.createElement(
        "a",
        { href: typeof href === "string" ? href : "#", ...props },
        children as React.ReactNode,
      ),
    usePathname: () => "/",
    useRouter: () => ({ replace: () => {}, push: () => {} }),
    redirect: () => {},
    getPathname: () => "/",
  };
});
