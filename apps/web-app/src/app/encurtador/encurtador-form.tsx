"use client";

import { useState, type FormEvent } from "react";

import { AnalyticsDashboard, type Analytics } from "./analytics-dashboard";
import styles from "./page.module.css";

interface CreatedLink {
  slug: string;
  short_url: string;
  original_url: string;
  expires_at: string | null;
}

type CreateStatus = "idle" | "sending" | "error";
type StatsStatus = "idle" | "loading" | "error";
type Copied = "short" | "analytics" | null;

// Caminho (same-origin) da página de análise de um slug.
function analyticsPath(slug: string): string {
  return `/encurtador/analise/${slug}`;
}

export function EncurtadorForm() {
  const [createStatus, setCreateStatus] = useState<CreateStatus>("idle");
  const [createError, setCreateError] = useState<string | null>(null);
  const [created, setCreated] = useState<CreatedLink | null>(null);
  const [copied, setCopied] = useState<Copied>(null);

  const [slugQuery, setSlugQuery] = useState("");
  const [statsStatus, setStatsStatus] = useState<StatsStatus>("idle");
  const [statsError, setStatsError] = useState<string | null>(null);
  const [analytics, setAnalytics] = useState<Analytics | null>(null);

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const originalUrl = String(data.get("original_url") ?? "").trim();
    const customAlias = String(data.get("custom_alias") ?? "").trim();
    const expireAt = String(data.get("expire_at") ?? "").trim();

    const payload: Record<string, string> = { original_url: originalUrl };
    if (customAlias) payload.custom_alias = customAlias;
    if (expireAt) payload.expire_at = new Date(expireAt).toISOString();

    setCreateStatus("sending");
    setCreateError(null);
    setCopied(null);

    try {
      const res = await fetch("/api/encurtador/urls", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const body = await res.json().catch(() => null);

      if (res.ok) {
        const link = body as CreatedLink;
        setCreated(link);
        setSlugQuery(link.slug);
        setAnalytics(null);
        setCreateStatus("idle");
        return;
      }
      setCreateError(
        body && typeof body.error === "string"
          ? body.error
          : "Não foi possível encurtar agora.",
      );
      setCreateStatus("error");
    } catch {
      setCreateError("Não foi possível encurtar agora.");
      setCreateStatus("error");
    }
  }

  async function loadAnalytics(slug: string) {
    const target = slug.trim();
    if (!target) return;

    setStatsStatus("loading");
    setStatsError(null);
    try {
      const res = await fetch(
        `/api/encurtador/analytics/${encodeURIComponent(target)}`,
      );
      const body = await res.json().catch(() => null);

      if (res.ok) {
        setAnalytics(body as Analytics);
        setStatsStatus("idle");
        return;
      }
      setAnalytics(null);
      setStatsError(
        res.status === 404
          ? "Nenhum link com esse código."
          : "Não foi possível carregar as métricas.",
      );
      setStatsStatus("error");
    } catch {
      setAnalytics(null);
      setStatsError("Não foi possível carregar as métricas.");
      setStatsStatus("error");
    }
  }

  async function copy(kind: "short" | "analytics", text: string) {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(kind);
    } catch {
      /* clipboard indisponível: ignora silenciosamente */
    }
  }

  const analyticsUrl = created
    ? `${typeof window !== "undefined" ? window.location.origin : ""}${analyticsPath(created.slug)}`
    : "";

  return (
    <>
      <form className={styles.form} onSubmit={handleCreate}>
        <div className={styles.field}>
          <label htmlFor="url-original">URL de destino</label>
          <input
            id="url-original"
            name="original_url"
            type="url"
            placeholder="https://exemplo.com/uma/pagina/bem/longa"
            required
          />
        </div>

        <div className={styles.row}>
          <div className={styles.field}>
            <label htmlFor="url-alias">Alias custom (opcional)</label>
            <input
              id="url-alias"
              name="custom_alias"
              type="text"
              placeholder="minha-marca"
              maxLength={40}
              pattern="[A-Za-z0-9_-]+"
            />
          </div>
          <div className={styles.field}>
            <label htmlFor="url-expira">Expira em (opcional)</label>
            <input id="url-expira" name="expire_at" type="datetime-local" />
          </div>
        </div>

        <div className={styles.actions}>
          <button
            type="submit"
            className={styles.submit}
            disabled={createStatus === "sending"}
          >
            {createStatus === "sending" ? "Encurtando…" : "Encurtar"}
          </button>
          <p className={styles.hint}>Slug base62 de 7 caracteres, ou o seu alias.</p>
        </div>

        <p className={styles.error} role="alert" aria-live="polite">
          {createStatus === "error" ? createError : ""}
        </p>
      </form>

      {created && (
        <div className={styles.result} role="status">
          <div className={styles.resultGroup}>
            <span className={styles.resultLabel}>Link curto</span>
            <div className={styles.resultRow}>
              <a
                className={styles.resultUrl}
                href={created.short_url}
                target="_blank"
                rel="noreferrer"
              >
                {created.short_url}
              </a>
              <button
                type="button"
                className={styles.copyButton}
                onClick={() => copy("short", created.short_url)}
              >
                {copied === "short" ? "copiado ✓" : "copiar"}
              </button>
            </div>
          </div>

          <div className={styles.resultGroup}>
            <span className={styles.resultLabel}>Link de análise</span>
            <div className={styles.resultRow}>
              <a
                className={styles.resultUrl}
                href={analyticsPath(created.slug)}
                target="_blank"
                rel="noreferrer"
              >
                {analyticsUrl || analyticsPath(created.slug)}
              </a>
              <button
                type="button"
                className={styles.copyButton}
                onClick={() =>
                  copy("analytics", analyticsUrl || analyticsPath(created.slug))
                }
              >
                {copied === "analytics" ? "copiado ✓" : "copiar"}
              </button>
            </div>
          </div>

          <p className={styles.resultOriginal}>→ {created.original_url}</p>
        </div>
      )}

      <section className={styles.analytics} aria-label="Métricas de clique">
        <h2 className={styles.analyticsTitle}>Métricas</h2>
        <div className={styles.lookup}>
          <input
            className={styles.lookupInput}
            value={slugQuery}
            onChange={(e) => setSlugQuery(e.target.value)}
            placeholder="código do link (ex: aB3x9)"
            aria-label="Código do link"
          />
          <button
            type="button"
            className={styles.lookupButton}
            onClick={() => loadAnalytics(slugQuery)}
            disabled={statsStatus === "loading" || slugQuery.trim() === ""}
          >
            {statsStatus === "loading" ? "Carregando…" : "Ver métricas"}
          </button>
        </div>

        {statsStatus === "error" && (
          <p className={styles.error} role="alert">
            {statsError}
          </p>
        )}

        {analytics && <AnalyticsDashboard analytics={analytics} />}
      </section>
    </>
  );
}
