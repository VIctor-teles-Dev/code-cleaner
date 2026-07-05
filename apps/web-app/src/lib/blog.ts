const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";

export interface Tag {
  slug: string;
  name: string;
}

export interface PostSummary {
  slug: string;
  title: string;
  excerpt: string;
  published_at: string | null;
  tags: Tag[];
}

export interface Post {
  slug: string;
  title: string;
  content: string;
  published_at: string | null;
  tags: Tag[];
}

// Fetch server-side direto no service interno; as páginas do blog são
// dinâmicas (cache: no-store) para novos posts aparecerem sem rebuild.
// O locale é repassado ao backend, que resolve título/conteúdo/tags por idioma.
export async function getPosts(
  locale: string,
  tag?: string,
): Promise<PostSummary[]> {
  const params = new URLSearchParams({ locale });
  if (tag) params.set("tag", tag);
  const response = await fetch(`${BACKEND_URL}/posts?${params}`, {
    cache: "no-store",
  });
  if (!response.ok) {
    throw new Error(`GET /posts failed: ${response.status}`);
  }
  return response.json();
}

export async function getPost(
  locale: string,
  slug: string,
): Promise<Post | null> {
  const params = new URLSearchParams({ locale });
  const response = await fetch(
    `${BACKEND_URL}/posts/${encodeURIComponent(slug)}?${params}`,
    { cache: "no-store" },
  );
  if (response.status === 404) {
    return null;
  }
  if (!response.ok) {
    throw new Error(`GET /posts/${slug} failed: ${response.status}`);
  }
  return response.json();
}

export function formatDate(iso: string | null, locale: string): string {
  if (!iso) {
    return "";
  }
  return new Intl.DateTimeFormat(locale, {
    day: "numeric",
    month: "long",
    year: "numeric",
    timeZone: "UTC",
  }).format(new Date(iso));
}
