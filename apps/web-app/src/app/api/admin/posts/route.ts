import { proxyToBackend } from "@/lib/admin";
import { requireApiSession } from "@/lib/auth";

export const runtime = "nodejs";

// GET  /api/admin/posts  -> lista todos os posts (inclui rascunhos)
export async function GET() {
  const denied = await requireApiSession();
  if (denied) return denied;
  return proxyToBackend("GET", "/admin/posts");
}

// POST /api/admin/posts  -> cria um post
export async function POST(request: Request) {
  const denied = await requireApiSession();
  if (denied) return denied;
  return proxyToBackend("POST", "/posts", await request.text());
}
