"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { NAV_ITEMS, isNavItemActive } from "./nav-items";
import styles from "./site-header.module.css";

export function NavLinks() {
  const pathname = usePathname();

  return (
    <nav className={styles.nav} aria-label="Navegação principal">
      {NAV_ITEMS.map(({ href, label }) => {
        const isActive = isNavItemActive(href, pathname);

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
