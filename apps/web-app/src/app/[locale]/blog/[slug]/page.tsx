import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getTranslations, setRequestLocale } from "next-intl/server";
import ReactMarkdown from "react-markdown";

import { Link } from "@/i18n/navigation";
import { formatDate, getPost } from "@/lib/blog";

import styles from "./page.module.css";

export const dynamic = "force-dynamic";

interface PostPageProps {
  params: Promise<{ locale: string; slug: string }>;
}

export async function generateMetadata({
  params,
}: PostPageProps): Promise<Metadata> {
  const { locale, slug } = await params;
  const post = await getPost(locale, slug);
  if (!post) {
    const t = await getTranslations({ locale, namespace: "blog" });
    return { title: t("notFound") };
  }
  return { title: post.title };
}

export default async function PostPage({ params }: PostPageProps) {
  const { locale, slug } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("blog");

  const post = await getPost(locale, slug);
  if (!post) {
    notFound();
  }

  return (
    <main className={styles.page}>
      <article className={styles.article}>
        <header className={styles.header}>
          <Link href="/blog" className={styles.back}>
            {t("back")}
          </Link>
          <time
            className={styles.date}
            dateTime={post.published_at ?? undefined}
          >
            {formatDate(post.published_at, locale)}
          </time>
          <h1 className={styles.title}>{post.title}</h1>
          {post.tags.length > 0 && (
            <ul className={styles.tags} aria-label={t("tagsAriaLabel")}>
              {post.tags.map((tag) => (
                <li key={tag.slug}>
                  <Link
                    href={`/blog?tag=${encodeURIComponent(tag.slug)}`}
                    className={styles.tag}
                  >
                    #{tag.name}
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </header>
        <div className={styles.prose}>
          <ReactMarkdown>{post.content}</ReactMarkdown>
        </div>
      </article>
    </main>
  );
}
