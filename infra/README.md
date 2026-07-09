# Infraestrutura

Docker + Kubernetes com roteamento por subdomínio:

| Host | Serviço |
| --- | --- |
| `code-cleaner.ccl.app.br` | web-app (página principal) |
| `api.code-cleaner.ccl.app.br` | backend-api |
| `<app>.ccl.app.br` | padrão para futuros apps — adicione uma regra no `k8s/ingress.yaml` |

O domínio raiz `ccl.app.br` fica reservado para um futuro hub/landing dos apps.

O PostgreSQL roda dentro do cluster (StatefulSet com volume persistente) e é
exposto internamente como `postgres:5432`. A connection string é injetada no
backend via `DATABASE_URL`, montada a partir do Secret `postgres-credentials`.
O backend valida a conexão no endpoint `/readyz` (readiness probe) — se o
banco cair, os pods saem do balanceamento sem serem reiniciados; `/healthz`
segue como liveness probe, sem dependências externas.

Na inicialização, o backend aplica as migrations embutidas no binário
(golang-migrate). É seguro com múltiplas réplicas: o driver serializa via
advisory lock do Postgres e um banco já migrado é no-op. Não há passo manual
de migração no deploy — basta subir a imagem nova.

## Docker Compose (dev local)

Na raiz do repositório:

```bash
docker compose up -d --build
```

| Serviço | URL/porta no host |
| --- | --- |
| web-app | http://localhost:3000 |
| backend-api | http://localhost:8081 (host 8081 → container 8080) |
| postgres | localhost:5432 (user/senha/db: `ccl`/`ccl`/`ccl`) |

## Kubernetes local (kind)

Pré-requisitos: [kind](https://kind.sigs.k8s.io) e [kubectl](https://kubernetes.io/docs/tasks/tools/).

> **Atenção:** o cluster mapeia as portas 80/443 do host. Se já houver algo
> nelas (ex.: um proxy de outro projeto), pare-o ou ajuste `hostPort` em
> `infra/kind/cluster.yaml`.

```bash
# 1. Crie o cluster
kind create cluster --name ccl --config infra/kind/cluster.yaml

# 2. Instale o ingress-nginx (variante para kind)
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
kubectl wait --namespace ingress-nginx --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller --timeout=120s

# 2b. Instale o sealed-secrets (controller que dessela os Secrets do git)
# Em cluster RECRIADO: restaure a chave de selagem ANTES do controller,
# senão os SealedSecrets do repo não abrem — veja a seção "Secrets".
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.38.4/controller.yaml
kubectl wait --namespace kube-system --for=condition=ready pod \
  -l name=sealed-secrets-controller --timeout=120s

# 3. Builde as imagens e carregue no cluster
docker build -t ccl/backend-api:dev apps/backend-api
docker build -f apps/web-app/Dockerfile -t ccl/web-app:dev .
kind load docker-image ccl/web-app:dev ccl/backend-api:dev --name ccl

# 4. Aplique os manifests — overlay de DEV (hosts *.localhost, HTTP, sem TLS).
#    NÃO use `infra/k8s` direto no kind: ele roteia pelos domínios REAIS de prod
#    e, com o /etc/hosts abaixo, o navegador para de enxergar o site no ar.
kubectl apply -k infra/overlays/local

# 5. Aponte os hosts de DEV para 127.0.0.1. Use *.localhost (não os *.ccl.app.br
#    de prod!). O IP explícito é de propósito: o kind só mapeia IPv4, e sem a
#    linha o *.localhost cai em ::1 (IPv6) e não conecta.
#    Se você já colou os *.ccl.app.br no /etc/hosts antes, REMOVA aquelas linhas.
echo "127.0.0.1 ccl.localhost code-cleaner.localhost api.code-cleaner.localhost url-shortener.localhost" | sudo tee -a /etc/hosts

# 6. Acesse (HTTP). Prod segue em *.ccl.app.br resolvendo pro VPS, sem colisão.
curl http://code-cleaner.localhost/api/health
curl http://api.code-cleaner.localhost/healthz   # liveness (processo no ar)
curl http://api.code-cleaner.localhost/readyz    # readiness (inclui ping no PostgreSQL)
```

> **Dev vs. prod:** o kind usa `infra/overlays/local` (hosts `*.localhost`, HTTP);
> a VPS usa `infra/k8s` direto (hosts `*.ccl.app.br`, HTTPS via cert-manager).
> Assim o cluster local nunca sombreia os domínios reais.

## Secrets (Sealed Secrets)

Secrets vivem no git **criptografados** como `SealedSecret`
(`k8s/postgres-sealed-secret.yaml`): o [kubeseal](https://github.com/bitnami-labs/sealed-secrets/releases)
criptografa com a chave pública do controller, e só o controller do cluster
consegue desselar. O `kubectl apply -k` já os aplica junto com o resto.

Para criar/atualizar um secret selado:

```bash
kubectl create secret generic meu-secret -n ccl \
  --from-literal=CHAVE=valor --dry-run=client -o yaml |
  kubeseal --format yaml > k8s/meu-secret-sealed.yaml
# adicione o arquivo em k8s/kustomization.yaml
```

> **Backup da chave de selagem (importante):** os SealedSecrets do repo só
> abrem no cluster que possui a chave privada. Faça backup fora do git:
>
> ```bash
> kubectl get secret -n kube-system -l sealedsecrets.bitnami.com/sealed-secrets-key \
>   -o yaml > ~/.config/ccl/sealed-secrets-key-backup.yaml && chmod 600 ~/.config/ccl/sealed-secrets-key-backup.yaml
> ```
>
> Ao recriar o cluster, restaure a chave **antes** de instalar o controller
> (`kubectl apply -f ~/.config/ccl/sealed-secrets-key-backup.yaml`); sem o
> backup, gere os SealedSecrets de novo com `kubeseal` no cluster novo.

## Email do formulário de contato

As mensagens do `/contato` são **sempre persistidas** na tabela
`contact_messages` do Postgres — nada se perde se o email falhar. A
notificação por email é opcional e sai do próprio backend (na VPS) via
relay SMTP com STARTTLS na porta 587.

> Por que um relay, e não um servidor SMTP próprio na VPS? Provedores de
> VPS bloqueiam a porta 25, IPs de VPS nascem em blocklists e, sem
> SPF/DKIM/PTR e reputação, a entrega direta cai em spam. O relay resolve
> isso e continua gratuito no volume de um formulário de contato
> (Gmail com app password ~500/dia; Brevo 300/dia).

Para habilitar, crie o Secret **selado** (exemplo com Gmail — gere uma
[senha de app](https://myaccount.google.com/apppasswords)):

```bash
kubectl create secret generic smtp-credentials -n ccl \
  --from-literal=SMTP_HOST=smtp.gmail.com \
  --from-literal=SMTP_PORT=587 \
  --from-literal=SMTP_USERNAME=seu-email@gmail.com \
  --from-literal=SMTP_PASSWORD=sua-senha-de-app \
  --from-literal=CONTACT_TO=seu-email@gmail.com \
  --dry-run=client -o yaml |
  kubeseal --format yaml > k8s/smtp-sealed-secret.yaml
# adicione em k8s/kustomization.yaml, depois:
kubectl apply -k .
kubectl rollout restart deploy/backend-api -n ccl
```

Sem o Secret, o backend loga `SMTP not configured` e segue apenas
gravando as mensagens no banco. Para lê-las:

```bash
kubectl exec -n ccl postgres-0 -- psql -U ccl -d ccl \
  -c "SELECT created_at, name, email, message FROM contact_messages ORDER BY id DESC LIMIT 10"
```

## Publicando no blog

Os posts vivem na tabela `posts` e saem pela API (`GET /posts`,
`GET /posts/{slug}`). A escrita é o `POST /posts`, protegido pelo token
`BLOG_ADMIN_TOKEN` (SealedSecret `blog-admin-token` no repo; o valor em
claro fica em `~/.config/ccl/blog-admin-token`). Conteúdo em Markdown:

```bash
TOKEN=$(cat ~/.config/ccl/blog-admin-token)
curl -X POST http://api.code-cleaner.ccl.app.br/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "meu-post",
    "title": "Meu post",
    "content": "## Olá\n\nTexto em **markdown**.",
    "published": true,
    "tags": ["minhas aplicações"]
  }'
```

`"published": false` cria um rascunho, que não aparece na listagem nem na
página até ser publicado (por ora, via SQL: `UPDATE posts SET published_at
= now() WHERE slug = '...'`).

## Produção (quando chegar a hora)

- DNS: registro wildcard `*.ccl.app.br` apontando para o load balancer do cluster.
- TLS: [cert-manager](https://cert-manager.io) com Let's Encrypt e a seção `tls` no Ingress.
- Secrets: já usamos sealed-secrets; em produção, gere um novo `postgres-sealed-secret.yaml` com senha forte selada pela chave do cluster de produção (os valores selados no repo são de desenvolvimento).
- Banco: avalie um Postgres gerenciado (RDS, Cloud SQL, Neon) em vez do StatefulSet.

## CD: deploy automático na VPS (k3s + runner self-hosted)

No merge para a `main`, o workflow `.github/workflows/deploy.yml` roda **depois** do CI
passar, em dois jobs:

- **`build`** (runner do GitHub `ubuntu-latest`): builda as 3 imagens e publica no
  **GHCR** (`ghcr.io/<owner>/ccl-*:<sha>`). O build fica FORA da VPS de propósito —
  buildar o Next.js numa VPS de 1.9 GB estourava a RAM (OOM) e derrubava o cluster.
- **`deploy`** (runner self-hosted na VPS): só **baixa** as imagens do GHCR, importa no
  containerd do k3s (`docker pull` → re-tag `ccl/<app>:dev` → `k3s ctr images import`),
  aplica os manifests (`k3s kubectl apply -k`) e faz `rollout`. O `deploy` autentica no
  GHCR com o `GITHUB_TOKEN` do próprio job (não precisa de secret no cluster).

As **migrations rodam sozinhas** no boot de cada serviço Go, então o rollout já aplica o
schema novo.

> A VPS de 1.9 GB roda com **1 réplica** por app (ver `replicas` nos manifests) e com
> **swap** (`/swapfile`, 4 GB) pra ter folga. Se migrar para um host maior, dá pra subir
> as réplicas de volta.

### Setup da VPS (uma vez)

```bash
# 1. k3s (sem o Traefik, já que o Ingress usa ingressClassName: nginx)
curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--disable traefik" sh -

# 2. ingress-nginx + sealed-secrets
sudo k3s kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml
sudo k3s kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.38.4/controller.yaml

# 3. Docker (para baixar as imagens do GHCR e importar no k3s — o build é no CI)
curl -fsSL https://get.docker.com | sh

# 4. Runner self-hosted do GitHub com a label `ccl-vps`
#    (repo > Settings > Actions > Runners > New self-hosted runner)
#    O usuário do runner precisa do grupo docker e de sudo SEM SENHA para o k3s:
sudo usermod -aG docker "$USER"
echo "$USER ALL=(ALL) NOPASSWD: $(command -v k3s)" | sudo tee /etc/sudoers.d/k3s-runner

# 5. Secrets: SealedSecrets são POR CLUSTER — os do repo foram selados para o
#    cluster local. Re-sele para a VPS (kubeseal contra a chave DESTE cluster) e
#    substitua os *-sealed-secret.yaml; o smtp fica fora do git (imperativo):
sudo k3s kubectl create namespace ccl
sudo k3s kubectl -n ccl create secret generic smtp-credentials \
  --from-literal=SMTP_HOST=smtp.gmail.com --from-literal=SMTP_PORT=587 \
  --from-literal=SMTP_USERNAME=voce@gmail.com --from-literal=SMTP_PASSWORD=app-password \
  --from-literal=CONTACT_FROM=voce@gmail.com --from-literal=CONTACT_TO=voce@gmail.com

# 6. DNS: aponte os subdomínios (code-cleaner, api.code-cleaner, curto) para o IP da VPS.
# 7. Swap (VPS pequena): fallocate -l 4G /swapfile && chmod 600 /swapfile &&
#    mkswap /swapfile && swapon /swapfile && echo '/swapfile none swap sw 0 0' >> /etc/fstab
# 8. Primeiro deploy: faça um push na main (o CD builda no GHCR e a VPS baixa).
#    Para rodar à mão: exporte SHA/OWNER/GHCR_USER/GHCR_TOKEN e chame bash infra/deploy.sh.
```

TLS: adicione cert-manager + Let's Encrypt e a seção `tls` no Ingress para HTTPS.
