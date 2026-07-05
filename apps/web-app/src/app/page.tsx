import Link from "next/link";

import styles from "./page.module.css";

const STACK = ["Go", "TypeScript", "Next.js", "PostgreSQL", "Kubernetes"];

export default function Home() {
  return (
    <main className={styles.hero}>
      <div className={styles.content}>
        <p className={styles.badge}>
          <span className={styles.badgeDot} aria-hidden="true" />
          Disponível para novos projetos
        </p>
        <p className={styles.eyebrow}>
          Olá, mundo! Eu sou o desenvolvedor por trás do
        </p>
        <h1 className={styles.title}>
          code-<span className={styles.highlight}>cleaner</span>
          <span className={styles.cursor} aria-hidden="true" />
        </h1>
        <p className={styles.subtitle}>
          Transformando café em código limpo, arquitetura escalável e
          aplicações que resolvem problemas reais.
        </p>
        <div className={styles.ctas}>
          <Link className={styles.primary} href="/aplicacoes">
            Ver projetos
          </Link>
          <Link className={styles.secondary} href="/blog">
            Ler o blog
            <span className={styles.arrow} aria-hidden="true">
              →
            </span>
          </Link>
        </div>
        <ul className={styles.stack} aria-label="Principais tecnologias">
          {STACK.map((tech) => (
            <li key={tech}>{tech}</li>
          ))}
        </ul>
      </div>
    </main>
  );
}
