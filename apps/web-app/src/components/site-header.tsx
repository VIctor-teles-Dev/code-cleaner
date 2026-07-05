import { getTranslations } from "next-intl/server";

import { Link } from "@/i18n/navigation";

import { LanguageSwitcher } from "./language-switcher";
import { Logo } from "./logo";
import { NavLinks } from "./nav-links";
import styles from "./site-header.module.css";

export async function SiteHeader() {
  const t = await getTranslations("nav");

  return (
    <header className={styles.header}>
      <div className={styles.inner}>
        <Link href="/" className={styles.brand} aria-label={t("brandAriaLabel")}>
          <Logo />
          <span className={styles.wordmark}>code-cleaner</span>
        </Link>
        <NavLinks />
        <LanguageSwitcher />
      </div>
    </header>
  );
}
