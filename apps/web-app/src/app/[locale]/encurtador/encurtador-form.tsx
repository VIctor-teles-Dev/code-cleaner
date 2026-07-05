"use client";

import { useLocale, useTranslations } from "next-intl";
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
type StatsErrorKey = "notFound" | "metricsError";

export function EncurtadorForm() {
  const t = useTranslations("shortener");
  const locale = useLocale();

  const [createStatus, setCreateStatus] = useState<CreateStatus>("idle");
  const [created, setCreated] = useState<CreatedLink | null>(null);
  const [copied, setCopied] = useState(false);

  const [slugQuery, setSlugQuery] = useState("");
  const [statsStatus, setStatsStatus] = useState<StatsStatus>("idle");
  const [statsErrorKey, setStatsErrorKey] = useState<StatsErrorKey>(
    "metricsError",
  );
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
    setCopied(false);

    try {
      const res = await fetch("/api/encurtador/urls", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        const link = (await res.json()) as CreatedLink;
        setCreated(link);
        setSlugQuery(link.slug);
        setAnalytics(null);
        setCreateStatus("idle");
        return;
      }
      setCreateStatus("error");
    } catch {
      setCreateStatus("error");
    }
  }

  async function loadAnalytics(slug: string) {
    const target = slug.trim();
    if (!target) return;

    setStatsStatus("loading");
    try {
      const res = await fetch(
        `/api/encurtador/analytics/${encodeURIComponent(target)}`,
      );

      if (res.ok) {
        setAnalytics((await res.json()) as Analytics);
        setStatsStatus("idle");
        return;
      }
      setAnalytics(null);
      setStatsErrorKey(res.status === 404 ? "notFound" : "metricsError");
      setStatsStatus("error");
    } catch {
      setAnalytics(null);
      setStatsErrorKey("metricsError");
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
          <label htmlFor="url-original">{t("urlLabel")}</label>
          <input
            id="url-original"
            name="original_url"
            type="url"
            placeholder={t("urlPlaceholder")}
            required
          />
        </div>

        <div className={styles.row}>
          <div className={styles.field}>
            <label htmlFor="url-alias">{t("aliasLabel")}</label>
            <input
              id="url-alias"
              name="custom_alias"
              type="text"
              placeholder={t("aliasPlaceholder")}
              maxLength={40}
              pattern="[A-Za-z0-9_-]+"
            />
          </div>
          <div className={styles.field}>
            <label htmlFor="url-expira">{t("expiresLabel")}</label>
            <input id="url-expira" name="expire_at" type="datetime-local" />
          </div>
        </div>

        <div className={styles.actions}>
          <button
            type="submit"
            className={styles.submit}
            disabled={createStatus === "sending"}
          >
            {createStatus === "sending" ? t("shortening") : t("submit")}
          </button>
          <p className={styles.hint}>{t("hint")}</p>
        </div>

        <p className={styles.error} role="alert" aria-live="polite">
          {createStatus === "error" ? t("createError") : ""}
        </p>
      </form>

      {created && (
        <div className={styles.result} role="status">
          <span className={styles.resultLabel}>{t("resultLabel")}</span>
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
              {copied ? t("copied") : t("copy")}
            </button>
          </div>
          <p className={styles.resultOriginal}>→ {created.original_url}</p>
        </div>
      )}

      <section className={styles.analytics} aria-label={t("metricsAriaLabel")}>
        <h2 className={styles.analyticsTitle}>{t("metricsHeading")}</h2>
        <div className={styles.lookup}>
          <input
            className={styles.lookupInput}
            value={slugQuery}
            onChange={(e) => setSlugQuery(e.target.value)}
            placeholder={t("slugPlaceholder")}
            aria-label={t("slugAriaLabel")}
          />
          <button
            type="button"
            className={styles.lookupButton}
            onClick={() => loadAnalytics(slugQuery)}
            disabled={statsStatus === "loading" || slugQuery.trim() === ""}
          >
            {statsStatus === "loading" ? t("loading") : t("viewMetrics")}
          </button>
        </div>

        {statsStatus === "error" && (
          <p className={styles.error} role="alert">
            {t(statsErrorKey)}
          </p>
        )}

        {analytics && (
          <div className={styles.dashboard}>
            <div className={styles.total}>
              <span className={styles.totalNumber}>
                {analytics.total_clicks}
              </span>
              <span className={styles.totalLabel}>
                {t("clicksLabel", { count: analytics.total_clicks })}{" "}
                <code>{analytics.slug}</code>
              </span>
            </div>

            <div className={styles.panels}>
              <BarList
                title={t("panelDay")}
                items={analytics.time_series.map((d) => ({
                  label: formatDay(d.day, locale),
                  count: d.count,
                }))}
                empty={t("emptyClicks")}
              />
              <BarList
                title={t("panelCountries")}
                items={analytics.top_countries}
                empty={t("emptyCountries")}
              />
              <BarList
                title={t("panelBrowsers")}
                items={analytics.browsers}
                empty={t("emptyData")}
              />
              <BarList
                title={t("panelDevices")}
                items={analytics.devices}
                empty={t("emptyData")}
              />
              <BarList
                title={t("panelReferrers")}
                items={analytics.top_referrers}
                empty={t("emptyReferrers")}
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

function formatDay(iso: string, locale: string): string {
  return new Intl.DateTimeFormat(locale, {
    day: "2-digit",
    month: "2-digit",
    timeZone: "UTC",
  }).format(new Date(iso));
}
