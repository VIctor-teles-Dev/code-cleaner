import createMiddleware from "next-intl/middleware";

import { routing } from "./i18n/routing";

export default createMiddleware(routing);

export const config = {
  // Ignora API, assets internos e arquivos com extensão (favicon, ícones, etc.).
  matcher: ["/((?!api|_next|_vercel|.*\\..*).*)"],
};
