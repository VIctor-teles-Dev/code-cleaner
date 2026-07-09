import { proxyToBackend } from "@/lib/admin";
import { requireApiSession } from "@/lib/auth";

export const runtime = "nodejs";

type Context = { params: Promise<{ slug: string }> };

// GET /api/admin/posts/{slug} -> post completo p/ edição (inclui rascunho)
export async function GET(_request: Request, { params }: Context) {
  const denied = await requireApiSession();
  if (denied) return denied;
  const { slug } = await params;
  return proxyToBackend("GET", `/admin/posts/${encodeURIComponent(slug)}`);
}

// PUT /api/admin/posts/{slug} -> edita o post
export async function PUT(request: Request, { params }: Context) {
  const denied = await requireApiSession();
  if (denied) return denied;
  const { slug } = await params;
  return proxyToBackend(
    "PUT",
    `/posts/${encodeURIComponent(slug)}`,
    await request.text(),
  );
}

// DELETE /api/admin/posts/{slug} -> remove o post
export async function DELETE(_request: Request, { params }: Context) {
  const denied = await requireApiSession();
  if (denied) return denied;
  const { slug } = await params;
  return proxyToBackend("DELETE", `/posts/${encodeURIComponent(slug)}`);
}
