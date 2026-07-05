import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import ReactMarkdown from "react-markdown";

import { formatDate, getPost } from "@/lib/blog";

import styles from "./page.module.css";

export const dynamic = "force-dynamic";

interface PostPageProps {
  params: Promise<{ slug: string }>;
}

export async function generateMetadata({
  params,
}: PostPageProps): Promise<Metadata> {
  const { slug } = await params;
  const post = await getPost(slug);
  if (!post) {
    return { title: "Post não encontrado" };
  }
  return { title: post.title };
}

export default async function PostPage({ params }: PostPageProps) {
  const { slug } = await params;
  const post = await getPost(slug);
  if (!post) {
    notFound();
  }

  return (
    <main className={styles.page}>
      <article className={styles.article}>
        <header className={styles.header}>
          <Link href="/blog" className={styles.back}>
            ← Voltar para o blog
          </Link>
          <time
            className={styles.date}
            dateTime={post.published_at ?? undefined}
          >
            {formatDate(post.published_at)}
          </time>
          <h1 className={styles.title}>{post.title}</h1>
          {post.tags.length > 0 && (
            <ul className={styles.tags} aria-label="Tags">
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
