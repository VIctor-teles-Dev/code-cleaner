import { createNavigation } from "next-intl/navigation";

import { routing } from "./routing";

// Wrappers com locale automático: Link/usePathname/useRouter já lidam com o
// prefixo de idioma (usePathname retorna o caminho SEM o locale).
export const { Link, redirect, usePathname, useRouter, getPathname } =
  createNavigation(routing);
