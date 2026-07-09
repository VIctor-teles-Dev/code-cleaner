// Acesso server-side ao backend do blog com o BLOG_ADMIN_TOKEN. Usado pelas
// páginas do admin (server components) e pelos proxies /api/admin/* — o token
// nunca chega ao navegador.
import type { Post, PostSummary } from "@/lib/blog";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";
const BLOG_TOKEN = process.env.BLOG_ADMIN_TOKEN ?? "";

function authHeaders(): Record<string, string> {
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${BLOG_TOKEN}`,
  };
}

// Lista todos os posts, inclusive rascunhos (server component da lista).
export async function listAllPosts(): Promise<PostSummary[]> {
  const res = await fetch(`${BACKEND_URL}/admin/posts`, {
    headers: authHeaders(),
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`GET /admin/posts failed: ${res.status}`);
  }
  return res.json();
}

// Busca um post (publicado ou rascunho) para edição.
export async function getAnyPost(slug: string): Promise<Post | null> {
  const res = await fetch(
    `${BACKEND_URL}/admin/posts/${encodeURIComponent(slug)}`,
    { headers: authHeaders(), cache: "no-store" },
  );
  if (res.status === 404) {
    return null;
  }
  if (!res.ok) {
    throw new Error(`GET /admin/posts/${slug} failed: ${res.status}`);
  }
  return res.json();
}

// Encaminha uma requisição autenticada ao backend injetando o Bearer token,
// repassando status e corpo. Usado pelos route handlers /api/admin/*.
export async function proxyToBackend(
  method: string,
  path: string,
  body?: string,
): Promise<Response> {
  try {
    const upstream = await fetch(`${BACKEND_URL}${path}`, {
      method,
      headers: authHeaders(),
      body,
      signal: AbortSignal.timeout(10_000),
    });
    const text = await upstream.text();
    return new Response(text || null, {
      status: upstream.status,
      headers: {
        "Content-Type": upstream.headers.get("Content-Type") ?? "application/json",
      },
    });
  } catch {
    return Response.json({ status: "error" }, { status: 502 });
  }
}
