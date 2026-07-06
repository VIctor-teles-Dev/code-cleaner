"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { NAV_ITEMS, isNavItemActive } from "./nav-items";
import styles from "./mobile-tab-bar.module.css";

// Ícones (stroke, currentColor) por rota. Feather-style, 24x24.
const ICONS: Record<string, React.ReactNode> = {
  "/": (
    <path d="M3 10.5 12 3l9 7.5M5.5 9.5V20a1 1 0 0 0 1 1H10v-6h4v6h3.5a1 1 0 0 0 1-1V9.5" />
  ),
  "/sobre": (
    <>
      <circle cx="12" cy="8" r="3.6" />
      <path d="M4.5 20.5c0-3.6 3.4-5.5 7.5-5.5s7.5 1.9 7.5 5.5" />
    </>
  ),
  "/aplicacoes": (
    <>
      <rect x="3.5" y="3.5" width="7" height="7" rx="1.5" />
      <rect x="13.5" y="3.5" width="7" height="7" rx="1.5" />
      <rect x="3.5" y="13.5" width="7" height="7" rx="1.5" />
      <rect x="13.5" y="13.5" width="7" height="7" rx="1.5" />
    </>
  ),
  "/blog": (
    <>
      <path d="M6 3.5h8l4 4V20a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V4.5a1 1 0 0 1 1-1Z" />
      <path d="M13.5 3.5V8H18M8.5 12.5h7M8.5 16h7" />
    </>
  ),
  "/contato": (
    <>
      <rect x="3.5" y="5" width="17" height="14" rx="2" />
      <path d="m4 7 8 5.5L20 7" />
    </>
  ),
};

export function MobileTabBar() {
  const pathname = usePathname();

  return (
    <nav className={styles.bar} aria-label="Navegação">
      {NAV_ITEMS.map(({ href, label }) => {
        const active = isNavItemActive(href, pathname);

        return (
          <Link
            key={href}
            href={href}
            aria-current={active ? "page" : undefined}
            className={active ? `${styles.tab} ${styles.active}` : styles.tab}
          >
            <svg
              className={styles.icon}
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.7"
              strokeLinecap="round"
              strokeLinejoin="round"
              aria-hidden="true"
            >
              {ICONS[href]}
            </svg>
            <span className={styles.label}>{label}</span>
          </Link>
        );
      })}
    </nav>
  );
}
