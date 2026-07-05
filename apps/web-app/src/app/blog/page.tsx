import type { Metadata } from "next";
import Link from "next/link";

import { formatDate, getPosts } from "@/lib/blog";

import styles from "./page.module.css";

export const metadata: Metadata = {
  title: "Blog",
  description:
    "Artigos sobre código limpo, arquitetura e o que aprendo construindo aplicações reais.",
};

export const dynamic = "force-dynamic";

interface BlogProps {
  searchParams?: Promise<{ tag?: string }>;
}

export default async function Blog({ searchParams }: BlogProps) {
  const { tag } = (await searchParams) ?? {};
  const posts = await getPosts(tag);
  const activeTag = tag
    ? posts.flatMap((p) => p.tags).find((t) => t.slug === tag)
    : undefined;

  return (
    <main className={styles.page}>
      <div className={styles.container}>
        <p className={styles.eyebrow}>
          {tag ? `$ ls blog/ | grep "${tag}"` : "$ ls blog/"}
        </p>
        <h1 className={styles.title}>Blog</h1>
        <p className={styles.subtitle}>
          O que eu aprendo construindo — código limpo, arquitetura e as
          decisões por trás deste site.
        </p>

        {tag && (
          <p className={styles.filter}>
            Filtrando por{" "}
            <span className={styles.filterTag}>{activeTag?.name ?? tag}</span>
            <Link href="/blog" className={styles.filterClear}>
              limpar filtro ×
            </Link>
          </p>
        )}

        {posts.length === 0 ? (
          <p className={styles.empty}>
            {tag
              ? "// nenhum post com essa tag"
              : "// nenhum post publicado ainda — volte em breve"}
          </p>
        ) : (
          <ul className={styles.list}>
            {posts.map((post) => (
              <li key={post.slug} className={styles.card}>
                <time
                  className={styles.date}
                  dateTime={post.published_at ?? undefined}
                >
                  {formatDate(post.published_at)}
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
                  <ul className={styles.tags} aria-label="Tags">
                    {post.tags.map((t) => (
                      <li key={t.slug}>
                        <Link
                          href={`/blog?tag=${encodeURIComponent(t.slug)}`}
                          className={styles.tag}
                        >
                          #{t.name}
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
