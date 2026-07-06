import type { Metadata } from "next";

import styles from "./page.module.css";

export const metadata: Metadata = {
  title: "Aplicações",
  description:
    "As aplicações do monorepo, cada uma no seu subdomínio de ccl.app.br.",
};

interface AppEntry {
  name: string;
  description: string;
  stack: string[];
  url?: string;
}

// Cada app do monorepo ganha um subdomínio em ccl.app.br via regra de
// Ingress (infra/k8s/ingress.yaml). Sem url = ainda não publicado.
const APPS: AppEntry[] = [
  {
    name: "url-shortener",
    description:
      "Encurtador de URLs com redirecionamento de baixa latência e analytics de clique. Backend em Go + PostgreSQL; a UI (criar link + painel de métricas) roda aqui no site.",
    stack: ["Go", "PostgreSQL", "Kubernetes"],
    url: "/encurtador",
  },
];

export default function Aplicacoes() {
  return (
    <main className={styles.page}>
      <div className={styles.container}>
        <p className={styles.eyebrow}>$ ls aplicacoes/</p>
        <h1 className={styles.title}>Aplicações</h1>
        <p className={styles.subtitle}>
          Os apps deste monorepo, cada um servido no seu próprio subdomínio de{" "}
          <code className={styles.domain}>ccl.app.br</code> pelo mesmo cluster
          Kubernetes.
        </p>

        <ul className={styles.grid}>
          {APPS.map((app) => {
            const online = Boolean(app.url);
            const body = (
              <>
                <div className={styles.cardHeader}>
                  <span
                    className={online ? styles.dotOnline : styles.dotSoon}
                    aria-hidden="true"
                  />
                  <h2 className={styles.appName}>{app.name}</h2>
                  <span className={styles.status}>
                    {online ? "online" : "em breve"}
                  </span>
                </div>
                <p className={styles.description}>{app.description}</p>
                <ul className={styles.stack} aria-label="Stack">
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
                {online ? (
                  <a href={app.url} className={styles.card}>
                    {body}
                    <span className={styles.open} aria-hidden="true">
                      {app.url?.replace(/^https?:\/\//, "")} ↗
                    </span>
                  </a>
                ) : (
                  <div className={`${styles.card} ${styles.cardSoon}`}>
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
