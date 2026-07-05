"use client";

import { useLocale, useTranslations } from "next-intl";

import { usePathname, useRouter } from "@/i18n/navigation";
import { routing } from "@/i18n/routing";

import styles from "./site-header.module.css";

export function LanguageSwitcher() {
  const locale = useLocale();
  const pathname = usePathname();
  const router = useRouter();
  const t = useTranslations("languageSwitcher");

  return (
    <div className={styles.langSwitch} role="group" aria-label={t("ariaLabel")}>
      {routing.locales.map((loc) => (
        <button
          key={loc}
          type="button"
          aria-current={loc === locale ? "true" : undefined}
          className={
            loc === locale
              ? `${styles.langButton} ${styles.langActive}`
              : styles.langButton
          }
          onClick={() => router.replace(pathname, { locale: loc })}
        >
          {t(loc)}
        </button>
      ))}
    </div>
  );
}
