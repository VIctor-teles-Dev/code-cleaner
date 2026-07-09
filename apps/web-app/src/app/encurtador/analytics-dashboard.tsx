import styles from "./page.module.css";

export interface LabelCount {
  label: string;
  count: number;
}

export interface DayCount {
  day: string;
  count: number;
}

export interface Analytics {
  slug: string;
  total_clicks: number;
  time_series: DayCount[];
  top_countries: LabelCount[];
  top_referrers: LabelCount[];
  browsers: LabelCount[];
  devices: LabelCount[];
}

// Painel de métricas de um slug. Compartilhado entre o formulário (busca manual)
// e a página /encurtador/analise/[slug] (link de análise de cada link criado).
export function AnalyticsDashboard({ analytics }: { analytics: Analytics }) {
  return (
    <div className={styles.dashboard}>
      <div className={styles.total}>
        <span className={styles.totalNumber}>{analytics.total_clicks}</span>
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
        <BarList title="Navegadores" items={analytics.browsers} empty="Sem dados." />
        <BarList title="Dispositivos" items={analytics.devices} empty="Sem dados." />
        <BarList
          title="Referrers"
          items={analytics.top_referrers}
          empty="Acessos diretos."
        />
      </div>
    </div>
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
