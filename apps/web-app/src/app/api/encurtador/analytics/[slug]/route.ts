const SHORTENER_URL = process.env.SHORTENER_URL ?? "http://localhost:8082";
const ADMIN_TOKEN = process.env.URL_SHORTENER_ADMIN_TOKEN ?? "";

// Proxy same-origin das métricas. params é assíncrono no Next 16.
export async function GET(
  _request: Request,
  { params }: { params: Promise<{ slug: string }> },
) {
  const { slug } = await params;

  try {
    const upstream = await fetch(
      `${SHORTENER_URL}/api/v1/analytics/${encodeURIComponent(slug)}`,
      {
        headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
        signal: AbortSignal.timeout(10_000),
      },
    );

    return Response.json(await upstream.json(), { status: upstream.status });
  } catch {
    return Response.json({ status: "error" }, { status: 502 });
  }
}
