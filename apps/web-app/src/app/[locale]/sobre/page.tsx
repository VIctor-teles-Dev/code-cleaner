import type { Metadata } from "next";
import Image from "next/image";
import { getTranslations, setRequestLocale } from "next-intl/server";

import styles from "./page.module.css";

// Os nomes são próprios (não traduzidos); só o "role" vem das mensagens.
const TECHS = [
  { name: "Go", roleKey: "go" },
  { name: "TypeScript", roleKey: "typescript" },
  { name: "React / Next.js", roleKey: "react" },
  { name: "PostgreSQL", roleKey: "postgresql" },
  { name: "Docker", roleKey: "docker" },
  { name: "Kubernetes", roleKey: "kubernetes" },
  { name: "Bun + Turborepo", roleKey: "monorepo" },
  { name: "GitHub Actions", roleKey: "ci" },
] as const;

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "about" });
  return { title: t("title"), description: t("description") };
}

export default async function Sobre({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("about");

  return (
    <main className={styles.page}>
      <section className={styles.intro}>
        <div className={styles.photoFrame}>
          <Image
            src="/foto-perfil.jpg"
            alt={t("photoAlt")}
            fill
            sizes="(max-width: 820px) 280px, 320px"
            className={styles.photo}
            priority
          />
        </div>
        <div className={styles.bio}>
          <p className={styles.eyebrow}>{t("eyebrow")}</p>
          <h1 className={styles.title}>{t("heading")}</h1>
          <p>{t("bio1")}</p>
          <p>{t("bio2")}</p>
        </div>
      </section>

      <section className={styles.techSection} aria-labelledby="techs-title">
        <h2 id="techs-title" className={styles.techTitle}>
          {t("techHeading")}
        </h2>
        <ul className={styles.techGrid}>
          {TECHS.map(({ name, roleKey }) => (
            <li key={name} className={styles.techCard}>
              <span className={styles.techName}>{name}</span>
              <span className={styles.techRole}>{t(`roles.${roleKey}`)}</span>
            </li>
          ))}
        </ul>
      </section>
    </main>
  );
}
