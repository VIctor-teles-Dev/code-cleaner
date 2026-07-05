# Infraestrutura

Docker + Kubernetes com roteamento por subdomínio:

| Host | Serviço |
| --- | --- |
| `wbc.app.br` | web-app (página principal) |
| `api.wbc.app.br` | backend-api |
| `app1.wbc.app.br` | padrão para futuros apps — adicione uma regra no `k8s/ingress.yaml` |

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
| postgres | localhost:5432 (user/senha/db: `wbc`/`wbc`/`wbc`) |

## Kubernetes local (kind)

Pré-requisitos: [kind](https://kind.sigs.k8s.io) e [kubectl](https://kubernetes.io/docs/tasks/tools/).

> **Atenção:** o cluster mapeia as portas 80/443 do host. Se já houver algo
> nelas (ex.: um proxy de outro projeto), pare-o ou ajuste `hostPort` em
> `infra/kind/cluster.yaml`.

```bash
# 1. Crie o cluster
kind create cluster --name wbc --config infra/kind/cluster.yaml

# 2. Instale o ingress-nginx (variante para kind)
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
kubectl wait --namespace ingress-nginx --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller --timeout=120s

# 3. Builde as imagens e carregue no cluster
docker build -t wbc/backend-api:dev apps/backend-api
docker build -f apps/web-app/Dockerfile -t wbc/web-app:dev .
kind load docker-image wbc/web-app:dev wbc/backend-api:dev --name wbc

# 4. Aplique os manifests
kubectl apply -k infra/k8s

# 5. Aponte os domínios para localhost
echo "127.0.0.1 wbc.app.br api.wbc.app.br" | sudo tee -a /etc/hosts

# 6. Acesse
curl http://wbc.app.br/api/health
curl http://api.wbc.app.br/healthz   # liveness (processo no ar)
curl http://api.wbc.app.br/readyz    # readiness (inclui ping no PostgreSQL)
```

## Produção (quando chegar a hora)

- DNS: registro wildcard `*.wbc.app.br` apontando para o load balancer do cluster.
- TLS: [cert-manager](https://cert-manager.io) com Let's Encrypt e a seção `tls` no Ingress.
- Secrets: gere `postgres-credentials` fora do git (sealed-secrets, external-secrets ou o secret manager do provedor) — o valor em `k8s/postgres.yaml` é só para desenvolvimento.
- Banco: avalie um Postgres gerenciado (RDS, Cloud SQL, Neon) em vez do StatefulSet.
