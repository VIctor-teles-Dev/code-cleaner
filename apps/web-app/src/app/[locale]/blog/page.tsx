import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";

import { Link } from "@/i18n/navigation";
import { formatDate, getPosts } from "@/lib/blog";

import styles from "./page.module.css";

export const dynamic = "force-dynamic";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "blog" });
  return { title: t("title"), description: t("description") };
}

interface BlogProps {
  params: Promise<{ locale: string }>;
  searchParams?: Promise<{ tag?: string }>;
}

export default async function Blog({ params, searchParams }: BlogProps) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("blog");

  const { tag } = (await searchParams) ?? {};
  const posts = await getPosts(locale, tag);
  const activeTag = tag
    ? posts.flatMap((p) => p.tags).find((tg) => tg.slug === tag)
    : undefined;

  return (
    <main className={styles.page}>
      <div className={styles.container}>
        <p className={styles.eyebrow}>
          {tag ? t("eyebrowFiltered", { tag }) : t("eyebrow")}
        </p>
        <h1 className={styles.title}>{t("heading")}</h1>
        <p className={styles.subtitle}>{t("subtitle")}</p>

        {tag && (
          <p className={styles.filter}>
            {t("filteringBy")}
            <span className={styles.filterTag}>{activeTag?.name ?? tag}</span>
            <Link href="/blog" className={styles.filterClear}>
              {t("clearFilter")}
            </Link>
          </p>
        )}

        {posts.length === 0 ? (
          <p className={styles.empty}>{tag ? t("emptyTag") : t("emptyAll")}</p>
        ) : (
          <ul className={styles.list}>
            {posts.map((post) => (
              <li key={post.slug} className={styles.card}>
                <time
                  className={styles.date}
                  dateTime={post.published_at ?? undefined}
                >
                  {formatDate(post.published_at, locale)}
                </time>
                <h2 className={styles.postTitle}>
                  <Link
                    href={`/blog/${post.slug}`}
                    className={styles.postLink}
                  >
                    {post.title}
                  </Link>
                </h2>
                <p className={styles.excerpt}>{post.excerpt}</p>
                {post.tags.length > 0 && (
                  <ul className={styles.tags} aria-label={t("tagsAriaLabel")}>
                    {post.tags.map((tg) => (
                      <li key={tg.slug}>
                        <Link
                          href={`/blog?tag=${encodeURIComponent(tg.slug)}`}
                          className={styles.tag}
                        >
                          #{tg.name}
                        </Link>
                      </li>
                    ))}
                  </ul>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>
    </main>
  );
}
