"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import styles from "./site-header.module.css";

const LINKS = [
  { href: "/", label: "Início" },
  { href: "/sobre", label: "Sobre" },
  { href: "/aplicacoes", label: "Aplicações" },
  { href: "/blog", label: "Blog" },
  { href: "/contato", label: "Contato" },
] as const;

export function NavLinks() {
  const pathname = usePathname();

  return (
    <nav className={styles.nav} aria-label="Navegação principal">
      {LINKS.map(({ href, label }) => {
        const isActive =
          href === "/" ? pathname === "/" : pathname?.startsWith(href);

        return (
          <Link
            key={href}
            href={href}
            aria-current={isActive ? "page" : undefined}
            className={isActive ? `${styles.link} ${styles.active}` : styles.link}
          >
            {label}
          </Link>
        );
      })}
    </nav>
  );
}
