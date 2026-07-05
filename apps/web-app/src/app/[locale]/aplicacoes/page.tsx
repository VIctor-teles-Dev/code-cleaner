import type { Metadata } from "next";
import Image from "next/image";
import { getTranslations, setRequestLocale } from "next-intl/server";

import styles from "./page.module.css";

interface AppEntry {
  name: string;
  stack: string[];
  url?: string;
  screenshot?: string;
  building?: boolean;
}

// Cada app do monorepo ganha um subdomínio em ccl.app.br via regra de
// Ingress (infra/k8s/ingress.yaml). Sem url = ainda não publicado;
// building = em construção (preview com screenshot, sem link ainda).
// As descrições vêm das mensagens (apps.items.<name>).
const APPS: AppEntry[] = [
  {
    name: "code-cleaner",
    stack: ["Next.js", "Go", "PostgreSQL", "Kubernetes"],
    url: "http://code-cleaner.ccl.app.br",
  },
  {
    name: "url-shortener",
    stack: ["Go", "PostgreSQL", "Next.js", "Kubernetes"],
    screenshot: "/url-shortener.png",
    building: true,
  },
  {
    name: "mobile-app",
    stack: ["React Native", "Expo", "TypeScript"],
  },
  {
    name: "docs",
    stack: ["Next.js", "MDX"],
  },
];

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "apps" });
  return { title: t("title"), description: t("description") };
}

export default async function Aplicacoes({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("apps");

  return (
    <main className={styles.page}>
      <div className={styles.container}>
        <p className={styles.eyebrow}>{t("eyebrow")}</p>
        <h1 className={styles.title}>{t("heading")}</h1>
        <p className={styles.subtitle}>
          {t("subtitlePrefix")}
          <code className={styles.domain}>ccl.app.br</code>
          {t("subtitleSuffix")}
        </p>

        <ul className={styles.grid}>
          {APPS.map((app) => {
            const status = app.building
              ? "building"
              : app.url
                ? "online"
                : "soon";
            const dotClass =
              status === "online"
                ? styles.dotOnline
                : status === "building"
                  ? styles.dotBuilding
                  : styles.dotSoon;

            const body = (
              <>
                {app.screenshot && (
                  <div className={styles.thumb}>
                    <Image
                      src={app.screenshot}
                      alt={t("screenshotAlt", { name: app.name })}
                      fill
                      sizes="(max-width: 600px) 100vw, 340px"
                      className={styles.thumbImg}
                    />
                  </div>
                )}
                <div className={styles.cardHeader}>
                  <span className={dotClass} aria-hidden="true" />
                  <h2 className={styles.appName}>{app.name}</h2>
                  <span className={styles.status}>{t(`status.${status}`)}</span>
                </div>
                <p className={styles.description}>{t(`items.${app.name}`)}</p>
                <ul className={styles.stack} aria-label={t("stackAriaLabel")}>
                  {app.stack.map((tech) => (
                    <li key={tech} className={styles.tech}>
                      {tech}
                    </li>
                  ))}
                </ul>
              </>
            );

            return (
              <li key={app.name}>
                {status === "online" ? (
                  <a href={app.url} className={styles.card}>
                    {body}
                    <span className={styles.open} aria-hidden="true">
                      {app.url?.replace("http://", "")} ↗
                    </span>
                  </a>
                ) : (
                  <div
                    className={
                      status === "soon"
                        ? `${styles.card} ${styles.cardSoon}`
                        : styles.card
                    }
                  >
                    {body}
                  </div>
                )}
              </li>
            );
          })}
        </ul>
      </div>
    </main>
  );
}
