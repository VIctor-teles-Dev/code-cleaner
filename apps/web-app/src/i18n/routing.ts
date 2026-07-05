import { defineRouting } from "next-intl/routing";

// pt-BR é o padrão; ambos os locales aparecem na URL (/pt-BR, /en).
export const routing = defineRouting({
  locales: ["pt-BR", "en"],
  defaultLocale: "pt-BR",
  localePrefix: "always",
});

export type Locale = (typeof routing.locales)[number];
