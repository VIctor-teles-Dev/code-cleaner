# Build a partir da RAIZ do monorepo:
#   docker build -f apps/web-app/Dockerfile -t ccl/web-app:dev .

FROM oven/bun:1.3.9-alpine AS deps
WORKDIR /repo
COPY package.json bun.lock turbo.json ./
COPY apps/web-app/package.json apps/web-app/package.json
COPY apps/backend-api/package.json apps/backend-api/package.json
COPY apps/url-shortener/package.json apps/url-shortener/package.json
COPY packages/ui-components/package.json packages/ui-components/package.json
COPY packages/utils/package.json packages/utils/package.json
COPY packages/ts-config/package.json packages/ts-config/package.json
RUN bun install --frozen-lockfile

FROM oven/bun:1.3.9-alpine AS builder
WORKDIR /repo
COPY --from=deps /repo ./
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1
RUN bun run --cwd apps/web-app build

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    HOSTNAME=0.0.0.0 \
    PORT=3000
COPY --from=builder --chown=node:node /repo/apps/web-app/.next/standalone ./
COPY --from=builder --chown=node:node /repo/apps/web-app/.next/static ./apps/web-app/.next/static
COPY --from=builder --chown=node:node /repo/apps/web-app/public ./apps/web-app/public
USER node
EXPOSE 3000
CMD ["node", "apps/web-app/server.js"]
