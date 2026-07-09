"use client";

import { useEffect, useState } from "react";

import { AnalyticsDashboard, type Analytics } from "./analytics-dashboard";
import styles from "./page.module.css";

type Status = "loading" | "ready" | "not-found" | "error";

// Carrega e mostra as métricas de um slug via proxy same-origin.
// É o corpo do link de análise: /encurtador/analise/<slug>.
export function AnalyticsView({ slug }: { slug: string }) {
  const [status, setStatus] = useState<Status>("loading");
  const [analytics, setAnalytics] = useState<Analytics | null>(null);

  useEffect(() => {
    let active = true;

    async function load() {
      setStatus("loading");
      try {
        const res = await fetch(
          `/api/encurtador/analytics/${encodeURIComponent(slug)}`,
        );
        const body = await res.json().catch(() => null);
        if (!active) return;

        if (res.ok) {
          setAnalytics(body as Analytics);
          setStatus("ready");
          return;
        }
        setAnalytics(null);
        setStatus(res.status === 404 ? "not-found" : "error");
      } catch {
        if (!active) return;
        setAnalytics(null);
        setStatus("error");
      }
    }

    load();
    return () => {
      active = false;
    };
  }, [slug]);

  return (
    <>
      <p className={styles.eyebrow}>$ curl /api/v1/analytics/{slug}</p>
      <h1 className={styles.title}>Métricas do link</h1>
      <p className={styles.subtitle}>
        Análise de cliques do código <code>{slug}</code>. Cada acesso ao link
        curto alimenta este painel de forma assíncrona.
      </p>

      {status === "loading" && (
        <p className={styles.hint} role="status">
          Carregando métricas…
        </p>
      )}

      {status === "not-found" && (
        <p className={styles.error} role="alert">
          Nenhum link com esse código.
        </p>
      )}

      {status === "error" && (
        <p className={styles.error} role="alert">
          Não foi possível carregar as métricas.
        </p>
      )}

      {status === "ready" && analytics && (
        <AnalyticsDashboard analytics={analytics} />
      )}
    </>
  );
}
