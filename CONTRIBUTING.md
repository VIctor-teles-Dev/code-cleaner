# Fluxo de contribuição — tudo passa por Pull Request

A branch `main` é protegida: **não aceita push direto, force-push nem exclusão**.
Toda mudança na aplicação — código, infra ou documentação — entra por Pull Request.
Isso vale inclusive para quem tem acesso de admin no repositório.

## Por que PRs?

- **Histórico revisável** — cada mudança tem contexto, diff isolado e discussão.
- **`main` sempre íntegra** — é a branch que gera as imagens Docker e o deploy
  no cluster; ela nunca recebe código que não passou pelo fluxo.
- **Base para automação** — checks de CI (testes, lint, build) plugam direto no PR.

## Passo a passo

```bash
# 1. Parta da main atualizada
git checkout main && git pull

# 2. Crie uma branch com prefixo por tipo
git checkout -b feat/nome-da-feature      # ou fix/, docs/, infra/, refactor/

# 3. Desenvolva com TDD (red -> green -> refactor)
bun run test:watch    # no workspace em que estiver trabalhando

# 4. Antes de abrir o PR, rode a suíte completa na raiz
bun run test && bun run lint && bun run build

# 5. Commit e push
git add -A && git commit -m "feat: descreve a mudança"
git push -u origin feat/nome-da-feature

# 6. Abra o PR
gh pr create --fill    # ou pela interface do GitHub

# 7. Após aprovação/checks, faça o merge (squash) e apague a branch
gh pr merge --squash --delete-branch
```

## Convenções

| Item | Padrão |
| --- | --- |
| Branches | `feat/…`, `fix/…`, `docs/…`, `infra/…`, `refactor/…` |
| Commits | [Conventional Commits](https://www.conventionalcommits.org/pt-br/): `feat:`, `fix:`, `docs:`, `infra:`, `refactor:`, `test:` |
| Merge | Squash — um commit por PR na `main` |
| Testes | Obrigatórios para código novo (TDD, ver [README](README.md#fluxo-tdd)) |

## Depois do merge

Cada PR roda automaticamente o workflow de CI
([`.github/workflows/ci.yml`](.github/workflows/ci.yml)): testes, lint e build
de todos os workspaces (Vitest, ESLint, `tsc`, `go test`, `go vet`).

O deploy no cluster local (kind) é manual por enquanto — rebuild das imagens,
`kind load` e `kubectl apply`, conforme o [infra/README.md](infra/README.md).
