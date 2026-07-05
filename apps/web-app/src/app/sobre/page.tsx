import type { Metadata } from "next";
import Image from "next/image";

import styles from "./page.module.css";

export const metadata: Metadata = {
  title: "Sobre",
  description:
    "Quem está por trás do code-cleaner: objetivos, stack e forma de trabalhar.",
};

const TECHS = [
  { name: "Go", role: "APIs e serviços de backend" },
  { name: "TypeScript", role: "Tipagem forte do front ao back" },
  { name: "React / Next.js", role: "Interfaces modernas com SSR" },
  { name: "PostgreSQL", role: "Modelagem e persistência de dados" },
  { name: "Docker", role: "Ambientes reproduzíveis" },
  { name: "Kubernetes", role: "Orquestração e deploys resilientes" },
  { name: "Bun + Turborepo", role: "Monorepos com DX rápida" },
  { name: "GitHub Actions", role: "CI/CD em cada pull request" },
];

export default function Sobre() {
  return (
    <main className={styles.page}>
      <section className={styles.intro}>
        <div className={styles.photoFrame}>
          <Image
            src="/foto-perfil.jpg"
            alt="Foto de Victor Teles"
            fill
            sizes="(max-width: 820px) 280px, 320px"
            className={styles.photo}
            priority
          />
        </div>
        <div className={styles.bio}>
          <p className={styles.eyebrow}>$ whoami</p>
          <h1 className={styles.title}>Sobre mim</h1>
          <p>
            Olá! Eu sou o Victor, desenvolvedor full stack apaixonado por
            transformar problemas complexos em soluções simples. Trabalho do
            banco de dados à interface — backend em Go, frontend com React e
            Next.js — sempre guiado por testes e por código que a próxima
            pessoa consegue ler.
          </p>
          <p>
            Meu objetivo é construir aplicações que resolvem problemas reais,
            com arquitetura escalável e deploys previsíveis. Este site é meu
            laboratório: cada funcionalidade nasce com TDD, passa por CI e
            roda em Kubernetes — e o que aprendo no caminho vira artigo no
            blog.
          </p>
        </div>
      </section>

      <section className={styles.techSection} aria-labelledby="techs-title">
        <h2 id="techs-title" className={styles.techTitle}>
          Tecnologias com que trabalho
        </h2>
        <ul className={styles.techGrid}>
          {TECHS.map(({ name, role }) => (
            <li key={name} className={styles.techCard}>
              <span className={styles.techName}>{name}</span>
              <span className={styles.techRole}>{role}</span>
            </li>
          ))}
        </ul>
      </section>
    </main>
  );
}
