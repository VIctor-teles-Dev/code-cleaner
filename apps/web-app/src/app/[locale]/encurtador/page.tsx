import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";

import { EncurtadorForm } from "./encurtador-form";
import styles from "./page.module.css";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "shortener" });
  return { title: t("title"), description: t("description") };
}

export default async function Encurtador({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("shortener");

  return (
    <main className={styles.page}>
      <div className={styles.container}>
        <p className={styles.eyebrow}>{t("eyebrow")}</p>
        <h1 className={styles.title}>{t("heading")}</h1>
        <p className={styles.subtitle}>{t("subtitle")}</p>
        <EncurtadorForm />
      </div>
    </main>
  );
}
