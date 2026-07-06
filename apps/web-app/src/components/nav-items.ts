// Itens de navegação, compartilhados entre a navbar (desktop) e a tab bar (mobile).
export const NAV_ITEMS = [
  { href: "/", label: "Início" },
  { href: "/sobre", label: "Sobre" },
  { href: "/aplicacoes", label: "Aplicações" },
  { href: "/blog", label: "Blog" },
  { href: "/contato", label: "Contato" },
] as const;

// Um href está ativo se for exatamente "/" ou se o pathname começar com ele.
export function isNavItemActive(href: string, pathname: string | null): boolean {
  return href === "/" ? pathname === "/" : Boolean(pathname?.startsWith(href));
}
