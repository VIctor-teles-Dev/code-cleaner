import type { Metadata } from "next";

import { EncurtadorForm } from "./encurtador-form";
import styles from "./page.module.css";

export const metadata: Metadata = {
  title: "Encurtador",
  description:
    "Encurtador de URLs com redirecionamento de baixa latência e analytics de clique — API em Go rodando no mesmo cluster Kubernetes.",
};

export default function Encurtador() {
  return (
    <main className={styles.page}>
      <div className={styles.container}>
        <p className={styles.eyebrow}>$ curl -X POST /api/v1/urls</p>
        <h1 className={styles.title}>Encurtador de URL</h1>
        <p className={styles.subtitle}>
          Cole uma URL longa e receba um link curto. O redirecionamento é
          servido por uma API em Go (cache em memória para os links quentes) e
          cada clique alimenta o painel de métricas de forma assíncrona, sem
          atrasar o redirect.
        </p>
        <EncurtadorForm />
      </div>
    </main>
  );
}
