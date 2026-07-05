import { getTranslations, setRequestLocale } from "next-intl/server";

import { Link } from "@/i18n/navigation";

import styles from "./page.module.css";

const STACK = ["Go", "TypeScript", "Next.js", "PostgreSQL", "Kubernetes"];

export default async function Home({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("home");

  return (
    <main className={styles.hero}>
      <div className={styles.content}>
        <p className={styles.badge}>
          <span className={styles.badgeDot} aria-hidden="true" />
          {t("badge")}
        </p>
        <p className={styles.eyebrow}>{t("eyebrow")}</p>
        <h1 className={styles.title}>
          code-<span className={styles.highlight}>cleaner</span>
          <span className={styles.cursor} aria-hidden="true" />
        </h1>
        <p className={styles.subtitle}>{t("subtitle")}</p>
        <div className={styles.ctas}>
          <Link className={styles.primary} href="/aplicacoes">
            {t("ctaProjects")}
          </Link>
          <Link className={styles.secondary} href="/blog">
            {t("ctaBlog")}
            <span className={styles.arrow} aria-hidden="true">
              →
            </span>
          </Link>
        </div>
        <ul className={styles.stack} aria-label={t("stackAriaLabel")}>
          {STACK.map((tech) => (
            <li key={tech}>{tech}</li>
          ))}
        </ul>
      </div>
    </main>
  );
}
