"use client";

import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";

import type { Post } from "@/lib/blog";

import styles from "./admin.module.css";

type Status = "idle" | "saving" | "error";

interface PostEditorProps {
  // Ausente => criação; presente => edição (slug imutável).
  initial?: Post;
}

export function PostEditor({ initial }: PostEditorProps) {
  const router = useRouter();
  const editing = initial !== undefined;

  const [slug, setSlug] = useState(initial?.slug ?? "");
  const [title, setTitle] = useState(initial?.title ?? "");
  const [tags, setTags] = useState(
    initial?.tags.map((t) => t.name).join(", ") ?? "",
  );
  const [content, setContent] = useState(initial?.content ?? "");
  const [published, setPublished] = useState(initial?.published_at != null);

  const [status, setStatus] = useState<Status>("idle");
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setStatus("saving");
    setError(null);

    const payload = {
      slug: slug.trim(),
      title: title.trim(),
      content,
      published,
      tags: tags
        .split(",")
        .map((t) => t.trim())
        .filter(Boolean),
    };

    const url = editing
      ? `/api/admin/posts/${encodeURIComponent(initial.slug)}`
      : "/api/admin/posts";

    try {
      const res = await fetch(url, {
        method: editing ? "PUT" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        router.push("/admin");
        router.refresh();
        return;
      }
      const body = await res.json().catch(() => null);
      setError(
        body && typeof body.error === "string"
          ? body.error
          : "Não foi possível salvar.",
      );
      setStatus("error");
    } catch {
      setError("Não foi possível salvar.");
      setStatus("error");
    }
  }

  return (
    <form className={styles.form} onSubmit={handleSubmit}>
      <div className={styles.field}>
        <label htmlFor="post-slug">slug</label>
        <input
          id="post-slug"
          value={slug}
          onChange={(e) => setSlug(e.target.value)}
          placeholder="meu-post"
          pattern="[a-z0-9]+(-[a-z0-9]+)*"
          required
          disabled={editing}
          readOnly={editing}
        />
        <p className={styles.hint}>
          {editing
            ? "o slug é a URL do post e não muda depois de criado"
            : "minúsculas, números e hífens — vira a URL /blog/<slug>"}
        </p>
      </div>

      <div className={styles.field}>
        <label htmlFor="post-title">título</label>
        <input
          id="post-title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Meu post"
          required
        />
      </div>

      <div className={styles.field}>
        <label htmlFor="post-tags">tags</label>
        <input
          id="post-tags"
          value={tags}
          onChange={(e) => setTags(e.target.value)}
          placeholder="go, kubernetes, arquitetura"
        />
        <p className={styles.hint}>separadas por vírgula (opcional)</p>
      </div>

      <div className={styles.field}>
        <label htmlFor="post-content">conteúdo (markdown)</label>
        <textarea
          id="post-content"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          rows={18}
          placeholder={"## Seção\n\nTexto em **markdown**."}
          required
        />
      </div>

      <label className={styles.checkbox}>
        <input
          type="checkbox"
          checked={published}
          onChange={(e) => setPublished(e.target.checked)}
        />
        publicado (some da listagem pública quando desmarcado)
      </label>

      <div className={styles.actions}>
        <button
          type="submit"
          className={styles.primary}
          disabled={status === "saving"}
        >
          {status === "saving" ? "Salvando…" : editing ? "Salvar" : "Criar post"}
        </button>
        <button
          type="button"
          className={styles.secondary}
          onClick={() => router.push("/admin")}
        >
          Cancelar
        </button>
      </div>

      <p className={styles.error} role="alert" aria-live="polite">
        {status === "error" ? error : ""}
      </p>
    </form>
  );
}
