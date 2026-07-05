import type { Metadata } from "next";
import Image from "next/image";

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
  screenshot?: string;
  building?: boolean;
}

// Cada app do monorepo ganha um subdomínio em ccl.app.br via regra de
// Ingress (infra/k8s/ingress.yaml). Sem url = ainda não publicado;
// building = em construção (preview com screenshot, sem link ainda).
const APPS: AppEntry[] = [
  {
    name: "code-cleaner",
    description:
      "Este site: portfólio e blog com API própria em Go, posts em Markdown no PostgreSQL e formulário de contato com persistência primeiro.",
    stack: ["Next.js", "Go", "PostgreSQL", "Kubernetes"],
    url: "http://code-cleaner.ccl.app.br",
  },
  {
    name: "url-shortener",
    description:
      "Encurtador de URLs com redirect de baixa latência (cache em memória), analytics de clique assíncrono e alias custom. API em Go no mesmo cluster; links curtos em curto.ccl.app.br.",
    stack: ["Go", "PostgreSQL", "Next.js", "Kubernetes"],
    screenshot: "/url-shortener.png",
    building: true,
  },
  {
    name: "mobile-app",
    description:
      "Futuro app mobile do ecossistema, compartilhando os pacotes de UI e utilitários do monorepo.",
    stack: ["React Native", "Expo", "TypeScript"],
  },
  {
    name: "docs",
    description:
      "Futura central de documentação técnica dos projetos — arquitetura, decisões e guias de contribuição.",
    stack: ["Next.js", "MDX"],
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
            const statusLabel =
              status === "online"
                ? "online"
                : status === "building"
                  ? "em construção"
                  : "em breve";

            const body = (
              <>
                {app.screenshot && (
                  <div className={styles.thumb}>
                    <Image
                      src={app.screenshot}
                      alt={`Screenshot da aplicação ${app.name}`}
                      fill
                      sizes="(max-width: 600px) 100vw, 340px"
                      className={styles.thumbImg}
                    />
                  </div>
                )}
                <div className={styles.cardHeader}>
                  <span className={dotClass} aria-hidden="true" />
                  <h2 className={styles.appName}>{app.name}</h2>
                  <span className={styles.status}>{statusLabel}</span>
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
