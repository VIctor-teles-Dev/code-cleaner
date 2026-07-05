"use client";

import { useTranslations } from "next-intl";

import { Link, usePathname } from "@/i18n/navigation";

import styles from "./site-header.module.css";

// href é o caminho (sem locale); key aponta pra tradução em nav.*.
const LINKS = [
  { href: "/", key: "home" },
  { href: "/sobre", key: "about" },
  { href: "/aplicacoes", key: "apps" },
  { href: "/blog", key: "blog" },
  { href: "/contato", key: "contact" },
] as const;

export function NavLinks() {
  const pathname = usePathname(); // já vem sem o prefixo de locale
  const t = useTranslations("nav");

  return (
    <nav className={styles.nav} aria-label={t("ariaLabel")}>
      {LINKS.map(({ href, key }) => {
        const isActive =
          href === "/" ? pathname === "/" : pathname.startsWith(href);

        return (
          <Link
            key={href}
            href={href}
            aria-current={isActive ? "page" : undefined}
            className={
              isActive ? `${styles.link} ${styles.active}` : styles.link
            }
          >
            {t(key)}
          </Link>
        );
      })}
    </nav>
  );
}
