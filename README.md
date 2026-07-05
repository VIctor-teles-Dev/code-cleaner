# write-better-codes

Monorepo com frontend em Next.js (Bun + TypeScript) e backend em Go, orquestrado pelo Turborepo. O desenvolvimento segue TDD: escreva o teste primeiro, veja-o falhar, implemente, veja-o passar.

## Estrutura

```
write-better-codes/
├── apps/
│   ├── web-app/          # Frontend Next.js (App Router, TypeScript, Vitest)
│   ├── backend-api/      # API em Go (net/http, testes com httptest)
│   ├── mobile-app/       # Placeholder — futuro app React Native/Expo
│   └── docs/             # Placeholder — futura documentação
├── packages/
│   ├── ui-components/    # Componentes React compartilhados (testados com Vitest)
│   ├── utils/            # Funções utilitárias compartilhadas
│   └── ts-config/        # Configurações TypeScript compartilhadas
├── infra/
│   ├── k8s/              # Manifests Kubernetes (kustomize)
│   └── kind/             # Config de cluster local kind
├── docker-compose.yml    # Stack local: web + api + PostgreSQL
├── package.json          # Workspaces (Bun) e scripts globais
└── turbo.json            # Pipeline de tasks do Turborepo
```

## Contribuindo

A `main` é protegida — toda mudança entra por Pull Request. O fluxo completo
(branches, convenções, checklist) está em [CONTRIBUTING.md](CONTRIBUTING.md).

## Pré-requisitos

- [Bun](https://bun.sh) >= 1.3
- [Go](https://go.dev) >= 1.26

## Comandos

Na raiz do repositório:

| Comando | Descrição |
| --- | --- |
| `bun install` | Instala as dependências de todos os workspaces |
| `bun run dev` | Sobe web-app (`:3000`) e backend-api (`:8080`) em paralelo |
| `bun run test` | Roda os testes de todos os workspaces (Vitest + `go test`) |
| `bun run test:watch` | Vitest em modo watch — o ciclo do TDD |
| `bun run lint` | ESLint (web-app), `tsc --noEmit` (packages) e `go vet` (backend) |
| `bun run build` | Build de produção de todos os apps |

Para rodar em um workspace específico, use filtros do Turbo, por exemplo:

```bash
bunx turbo run test --filter=web-app
bunx turbo run dev --filter=backend-api
```

## Fluxo TDD

1. **Red** — escreva um teste que descreva o comportamento desejado e veja-o falhar
   (`bun run test:watch` no workspace em que estiver trabalhando).
2. **Green** — implemente o mínimo necessário para o teste passar.
3. **Refactor** — melhore o código mantendo os testes verdes.

Exemplos de referência já no repositório:

- `apps/web-app/src/app/page.test.tsx` — teste de componente React com Testing Library
- `apps/backend-api/internal/handler/health_test.go` — teste de handler HTTP com `httptest`
- `apps/backend-api/internal/handler/ready_test.go` — teste com dependência fake (interface `Pinger`)
- `packages/utils/src/slugify.test.ts` — teste unitário de função pura
- `packages/ui-components/src/button.test.tsx` — teste de componente compartilhado

## Docker e Kubernetes

A aplicação roda com roteamento por subdomínio: `wbc.app.br` (web-app),
`api.wbc.app.br` (backend-api) e `app1.wbc.app.br` para futuros apps.

```bash
# stack local completa (web :3000, api :8081, postgres :5432)
docker compose up -d --build
```

O deploy em Kubernetes (Ingress por subdomínio + PostgreSQL persistente) está
documentado em [`infra/README.md`](infra/README.md).

## Pacotes compartilhados

Os pacotes são consumidos direto do código-fonte TypeScript (sem etapa de build).
O `web-app` os declara em `transpilePackages` no `next.config.ts`:

```ts
import { Button } from "@write-better-codes/ui-components";
import { slugify } from "@write-better-codes/utils";
```
