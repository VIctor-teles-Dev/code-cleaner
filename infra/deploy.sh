#!/usr/bin/env bash
# Deploy no cluster k3s da VPS. As imagens já vêm BUILDADAS do GHCR (job "build"
# do .github/workflows/deploy.yml roda num runner do GitHub) — aqui só baixamos e
# importamos no k3s. Buildar na VPS (1.9 GB) estourava a RAM (OOM), por isso o
# build saiu daqui. Idempotente — pode rodar quantas vezes for preciso.
#
# Requisitos na VPS: docker, k3s, sudo sem senha p/ o k3s, e as variáveis
# SHA / OWNER / GHCR_USER / GHCR_TOKEN (injetadas pelo workflow de Deploy).
set -euo pipefail

cd "$(dirname "$0")/.."

: "${SHA:?SHA não definido (é injetado pelo workflow)}"
reg="ghcr.io/$(echo "${OWNER:-victor-teles-dev}" | tr '[:upper:]' '[:lower:]')"

echo ">> login no GHCR"
echo "${GHCR_TOKEN:?}" | docker login ghcr.io -u "${GHCR_USER:?}" --password-stdin

echo ">> pull das imagens do GHCR + import no containerd do k3s"
for img in backend-api url-shortener web-app; do
  docker pull "$reg/ccl-${img}:$SHA"
  # Re-tag para o nome local que os manifests usam (ccl/<app>:dev, IfNotPresent).
  docker tag "$reg/ccl-${img}:$SHA" "ccl/${img}:dev"
  docker save "ccl/${img}:dev" | sudo k3s ctr images import -
done

echo ">> aplica os manifests"
sudo k3s kubectl apply -k infra/k8s

echo ">> rollout (força pods novos com as imagens recém-importadas)"
for dep in backend-api url-shortener web-app; do
  sudo k3s kubectl -n ccl rollout restart "deployment/${dep}"
done
for dep in backend-api url-shortener web-app; do
  sudo k3s kubectl -n ccl rollout status "deployment/${dep}" --timeout=180s
done

echo ">> deploy concluído"
