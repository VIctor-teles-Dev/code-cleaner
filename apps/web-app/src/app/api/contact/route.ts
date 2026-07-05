const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";

// Proxy same-origin para o backend: o navegador nunca fala com a API
// diretamente, então não há CORS nem exposição do endpoint interno.
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
    const upstream = await fetch(`${BACKEND_URL}/contact`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      signal: AbortSignal.timeout(10_000),
    });

    return Response.json(await upstream.json(), { status: upstream.status });
  } catch {
    return Response.json({ status: "error" }, { status: 502 });
  }
}
