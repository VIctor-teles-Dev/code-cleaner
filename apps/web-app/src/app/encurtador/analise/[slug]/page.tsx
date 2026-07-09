import type { Metadata } from "next";
import Link from "next/link";

import { AnalyticsView } from "../../analytics-view";
import styles from "../../page.module.css";

interface AnalisePageProps {
  params: Promise<{ slug: string }>;
}

export async function generateMetadata({
  params,
}: AnalisePageProps): Promise<Metadata> {
  const { slug } = await params;
  return {
    title: `Métricas · ${slug}`,
    description: `Análise de cliques do link encurtado ${slug}.`,
  };
}

export default async function AnalisePage({ params }: AnalisePageProps) {
  const { slug } = await params;

  return (
    <main className={styles.page}>
      <div className={styles.container}>
        <Link href="/encurtador" className={styles.back}>
          ← Voltar para o encurtador
        </Link>
        <AnalyticsView slug={slug} />
      </div>
    </main>
  );
}
