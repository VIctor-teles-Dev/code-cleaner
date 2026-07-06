"use client";

import { useState, type FormEvent } from "react";

import styles from "./page.module.css";

interface CreatedLink {
  slug: string;
  short_url: string;
  original_url: string;
  expires_at: string | null;
}

interface LabelCount {
  label: string;
  count: number;
}

interface DayCount {
  day: string;
  count: number;
}

interface Analytics {
  slug: string;
  total_clicks: number;
  time_series: DayCount[];
  top_countries: LabelCount[];
  top_referrers: LabelCount[];
  browsers: LabelCount[];
  devices: LabelCount[];
}

type CreateStatus = "idle" | "sending" | "error";
type StatsStatus = "idle" | "loading" | "error";

export function EncurtadorForm() {
  const [createStatus, setCreateStatus] = useState<CreateStatus>("idle");
  const [createError, setCreateError] = useState<string | null>(null);
  const [created, setCreated] = useState<CreatedLink | null>(null);
  const [copied, setCopied] = useState(false);

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
    setCopied(false);

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

  async function copyShortUrl() {
    if (!created) return;
    try {
      await navigator.clipboard.writeText(created.short_url);
      setCopied(true);
    } catch {
      /* clipboard indisponível: ignora silenciosamente */
    }
  }

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
              onClick={copyShortUrl}
            >
              {copied ? "copiado ✓" : "copiar"}
            </button>
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

        {analytics && (
          <div className={styles.dashboard}>
            <div className={styles.total}>
              <span className={styles.totalNumber}>
                {analytics.total_clicks}
              </span>
              <span className={styles.totalLabel}>
                {analytics.total_clicks === 1 ? "clique" : "cliques"} em{" "}
                <code>{analytics.slug}</code>
              </span>
            </div>

            <div className={styles.panels}>
              <BarList
                title="Cliques por dia"
                items={analytics.time_series.map((d) => ({
                  label: formatDay(d.day),
                  count: d.count,
                }))}
                empty="Sem cliques ainda."
              />
              <BarList
                title="Países"
                items={analytics.top_countries}
                empty="Sem dados de país."
              />
              <BarList
                title="Navegadores"
                items={analytics.browsers}
                empty="Sem dados."
              />
              <BarList
                title="Dispositivos"
                items={analytics.devices}
                empty="Sem dados."
              />
              <BarList
                title="Referrers"
                items={analytics.top_referrers}
                empty="Acessos diretos."
              />
            </div>
          </div>
        )}
      </section>
    </>
  );
}

function BarList({
  title,
  items,
  empty,
}: {
  title: string;
  items: LabelCount[];
  empty: string;
}) {
  const max = items.reduce((m, i) => Math.max(m, i.count), 0) || 1;
  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>{title}</h3>
      {items.length === 0 ? (
        <p className={styles.panelEmpty}>{empty}</p>
      ) : (
        <ul className={styles.bars}>
          {items.map((item) => (
            <li key={item.label} className={styles.barRow}>
              <span className={styles.barLabel} title={item.label}>
                {item.label}
              </span>
              <span className={styles.barTrack}>
                <span
                  className={styles.barFill}
                  style={{ width: `${(item.count / max) * 100}%` }}
                />
              </span>
              <span className={styles.barCount}>{item.count}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function formatDay(iso: string): string {
  return new Intl.DateTimeFormat("pt-BR", {
    day: "2-digit",
    month: "2-digit",
    timeZone: "UTC",
  }).format(new Date(iso));
}
