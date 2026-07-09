import { isAuthConfigured, startSession, verifyPassword } from "@/lib/auth";

export const runtime = "nodejs";

// Rate-limit simples em memória (a app roda com 1 réplica). Trava um IP após
// muitas falhas seguidas, mitigando brute-force na senha.
const WINDOW_MS = 10 * 60 * 1000;
const MAX_FAILURES = 5;
const failures = new Map<string, { count: number; resetAt: number }>();

function clientIp(request: Request): string {
  const fwd = request.headers.get("x-forwarded-for");
  return fwd?.split(",")[0]?.trim() || "local";
}

function isBlocked(ip: string): boolean {
  const entry = failures.get(ip);
  return entry !== undefined && entry.resetAt > Date.now() && entry.count >= MAX_FAILURES;
}

function recordFailure(ip: string): void {
  const now = Date.now();
  const entry = failures.get(ip);
  if (!entry || entry.resetAt <= now) {
    failures.set(ip, { count: 1, resetAt: now + WINDOW_MS });
  } else {
    entry.count += 1;
  }
}

export async function POST(request: Request) {
  if (!isAuthConfigured()) {
    return Response.json(
      { error: "admin não configurado no servidor" },
      { status: 503 },
    );
  }

  const ip = clientIp(request);
  if (isBlocked(ip)) {
    return Response.json(
      { error: "muitas tentativas — tente novamente mais tarde" },
      { status: 429 },
    );
  }

  let password = "";
  try {
    const body = await request.json();
    password = typeof body?.password === "string" ? body.password : "";
  } catch {
    return Response.json({ error: "corpo inválido" }, { status: 400 });
  }

  if (!verifyPassword(password)) {
    recordFailure(ip);
    return Response.json({ error: "senha incorreta" }, { status: 401 });
  }

  failures.delete(ip);
  // Secure só quando servido por HTTPS (ingress seta x-forwarded-proto).
  await startSession(request.headers.get("x-forwarded-proto") === "https");
  return Response.json({ status: "ok" });
}
