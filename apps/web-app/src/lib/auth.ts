// Autenticação do admin (server-side apenas — usa node:crypto e next/headers).
//
// Duas fronteiras distintas (ver infra/README): a senha protege "humano → UI";
// o BLOG_ADMIN_TOKEN (usado em lib/admin.ts) protege "UI → API pública".
//
// - A senha nunca fica em claro: guardamos só o hash scrypt em ADMIN_PASSWORD_HASH.
// - A sessão é um cookie httpOnly assinado com HMAC (SESSION_SECRET); não carrega
//   a senha nem o token da API. Se vazar, expira e não expõe segredo de valor.
import { createHmac, scryptSync, timingSafeEqual } from "node:crypto";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

const SESSION_COOKIE = "ccl_admin_session";
const SESSION_TTL_SECONDS = 8 * 60 * 60; // 8h

const SESSION_SECRET = process.env.SESSION_SECRET ?? "";
// Formato: scrypt$<saltHex>$<hashHex> (ver scripts/hash-password.mjs).
const PASSWORD_HASH = process.env.ADMIN_PASSWORD_HASH ?? "";

/** Auth só funciona quando os dois segredos estão configurados. */
export function isAuthConfigured(): boolean {
  return SESSION_SECRET !== "" && PASSWORD_HASH !== "";
}

/** Compara a senha com o hash scrypt em tempo constante. */
export function verifyPassword(password: string): boolean {
  const [scheme, saltHex, hashHex] = PASSWORD_HASH.split("$");
  if (scheme !== "scrypt" || !saltHex || !hashHex || !SESSION_SECRET) {
    return false;
  }
  const expected = Buffer.from(hashHex, "hex");
  if (expected.length === 0) {
    return false;
  }
  let derived: Buffer;
  try {
    derived = scryptSync(password, Buffer.from(saltHex, "hex"), expected.length);
  } catch {
    return false;
  }
  return timingSafeEqual(expected, derived);
}

function sign(payload: string): string {
  return createHmac("sha256", SESSION_SECRET).update(payload).digest("base64url");
}

function createSessionToken(): string {
  const payload = Buffer.from(
    JSON.stringify({ exp: Math.floor(Date.now() / 1000) + SESSION_TTL_SECONDS }),
  ).toString("base64url");
  return `${payload}.${sign(payload)}`;
}

function verifySessionToken(token: string | undefined): boolean {
  if (!token || !SESSION_SECRET) {
    return false;
  }
  const dot = token.indexOf(".");
  if (dot < 0) {
    return false;
  }
  const payload = token.slice(0, dot);
  const provided = Buffer.from(token.slice(dot + 1));
  const expected = Buffer.from(sign(payload));
  if (provided.length !== expected.length || !timingSafeEqual(provided, expected)) {
    return false;
  }
  try {
    const { exp } = JSON.parse(Buffer.from(payload, "base64url").toString());
    return typeof exp === "number" && exp > Math.floor(Date.now() / 1000);
  } catch {
    return false;
  }
}

/** True quando a requisição atual traz um cookie de sessão válido. */
export async function isAuthenticated(): Promise<boolean> {
  const token = (await cookies()).get(SESSION_COOKIE)?.value;
  return verifySessionToken(token);
}

/**
 * Emite o cookie de sessão assinado (chamar em route handler / server action).
 * secure marca o cookie como HTTPS-only — derive do x-forwarded-proto do
 * ingress, para funcionar tanto em prod (https) quanto no kind local (http).
 */
export async function startSession(secure: boolean): Promise<void> {
  (await cookies()).set(SESSION_COOKIE, createSessionToken(), {
    httpOnly: true,
    secure,
    sameSite: "lax",
    path: "/",
    maxAge: SESSION_TTL_SECONDS,
  });
}

/** Remove o cookie de sessão (logout). */
export async function endSession(): Promise<void> {
  (await cookies()).delete(SESSION_COOKIE);
}

/** Guarda de página (server component): redireciona ao login se não autenticado. */
export async function requireSession(): Promise<void> {
  if (!(await isAuthenticated())) {
    redirect("/admin/login");
  }
}

/** Guarda de route handler: devolve 401 quando não autenticado, senão null. */
export async function requireApiSession(): Promise<Response | null> {
  if (!(await isAuthenticated())) {
    return Response.json({ error: "não autenticado" }, { status: 401 });
  }
  return null;
}
