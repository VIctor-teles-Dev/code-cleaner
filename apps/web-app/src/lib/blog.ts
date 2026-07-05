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
export async function getPosts(tag?: string): Promise<PostSummary[]> {
  const query = tag ? `?tag=${encodeURIComponent(tag)}` : "";
  const response = await fetch(`${BACKEND_URL}/posts${query}`, {
    cache: "no-store",
  });
  if (!response.ok) {
    throw new Error(`GET /posts failed: ${response.status}`);
  }
  return response.json();
}

export async function getPost(slug: string): Promise<Post | null> {
  const response = await fetch(
    `${BACKEND_URL}/posts/${encodeURIComponent(slug)}`,
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

export function formatDate(iso: string | null): string {
  if (!iso) {
    return "";
  }
  return new Intl.DateTimeFormat("pt-BR", {
    day: "numeric",
    month: "long",
    year: "numeric",
    timeZone: "UTC",
  }).format(new Date(iso));
}
