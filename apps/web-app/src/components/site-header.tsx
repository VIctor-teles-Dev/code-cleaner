import Link from "next/link";

import { Logo } from "./logo";
import { NavLinks } from "./nav-links";
import styles from "./site-header.module.css";

export function SiteHeader() {
  return (
    <header className={styles.header}>
      <div className={styles.inner}>
        <Link href="/" className={styles.brand} aria-label="code-cleaner — página inicial">
          <Logo />
          <span className={styles.wordmark}>code-cleaner</span>
        </Link>
        <NavLinks />
      </div>
    </header>
  );
}
