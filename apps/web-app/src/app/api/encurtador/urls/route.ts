const SHORTENER_URL = process.env.SHORTENER_URL ?? "http://localhost:8082";
const ADMIN_TOKEN = process.env.URL_SHORTENER_ADMIN_TOKEN ?? "";

// Proxy same-origin para o encurtador: injeta o admin token do lado do
// servidor, então o navegador nunca vê o token nem fala direto com a API.
export async function POST(request: Request) {
  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return Response.json(
      { status: "invalid", error: "corpo da requisição inválido" },
      { status: 400 },
    );
  }

  try {
    const upstream = await fetch(`${SHORTENER_URL}/api/v1/urls`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${ADMIN_TOKEN}`,
      },
      body: JSON.stringify(body),
      signal: AbortSignal.timeout(10_000),
    });

    return Response.json(await upstream.json(), { status: upstream.status });
  } catch {
    return Response.json({ status: "error" }, { status: 502 });
  }
}
