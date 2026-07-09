"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { formatDate, type Post, type PostSummary } from "@/lib/blog";

import styles from "./admin.module.css";

export function PostList({ posts }: { posts: PostSummary[] }) {
  const router = useRouter();
  const [pending, setPending] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function togglePublish(post: PostSummary) {
    setPending(post.slug);
    setError(null);
    try {
      const detailRes = await fetch(
        `/api/admin/posts/${encodeURIComponent(post.slug)}`,
      );
      if (!detailRes.ok) {
        throw new Error("load");
      }
      const detail: Post = await detailRes.json();

      const res = await fetch(
        `/api/admin/posts/${encodeURIComponent(post.slug)}`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            title: detail.title,
            content: detail.content,
            published: detail.published_at == null, // alterna
            tags: detail.tags.map((t) => t.name),
          }),
        },
      );
      if (!res.ok) {
        throw new Error("save");
      }
      router.refresh();
    } catch {
      setError("Não foi possível alterar a publicação.");
    } finally {
      setPending(null);
    }
  }

  async function remove(post: PostSummary) {
    if (
      !window.confirm(`Excluir "${post.title}"? Isso não pode ser desfeito.`)
    ) {
      return;
    }
    setPending(post.slug);
    setError(null);
    try {
      const res = await fetch(
        `/api/admin/posts/${encodeURIComponent(post.slug)}`,
        { method: "DELETE" },
      );
      if (!res.ok) {
        throw new Error("delete");
      }
      router.refresh();
    } catch {
      setError("Não foi possível excluir.");
    } finally {
      setPending(null);
    }
  }

  if (posts.length === 0) {
    return <p className={styles.empty}>{"// nenhum post ainda"}</p>;
  }

  return (
    <>
      {error && (
        <p className={styles.error} role="alert">
          {error}
        </p>
      )}
      <ul className={styles.list}>
        {posts.map((post) => {
          const isPublished = post.published_at != null;
          const busy = pending === post.slug;
          return (
            <li key={post.slug} className={styles.row}>
              <div className={styles.rowMain}>
                <span
                  className={`${styles.status} ${isPublished ? styles.statusPublished : styles.statusDraft}`}
                >
                  {isPublished ? "publicado" : "rascunho"}
                </span>
                <div className={styles.rowText}>
                  <span className={styles.rowTitle}>{post.title}</span>
                  <span className={styles.rowMeta}>
                    /{post.slug}
                    {isPublished && ` · ${formatDate(post.published_at)}`}
                  </span>
                </div>
              </div>
              <div className={styles.rowActions}>
                <Link
                  href={`/admin/posts/${post.slug}/edit`}
                  className={styles.action}
                >
                  editar
                </Link>
                <button
                  type="button"
                  className={styles.action}
                  onClick={() => togglePublish(post)}
                  disabled={busy}
                >
                  {isPublished ? "despublicar" : "publicar"}
                </button>
                <button
                  type="button"
                  className={`${styles.action} ${styles.danger}`}
                  onClick={() => remove(post)}
                  disabled={busy}
                >
                  excluir
                </button>
              </div>
            </li>
          );
        })}
      </ul>
    </>
  );
}
