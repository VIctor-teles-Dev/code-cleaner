#!/usr/bin/env bash
# Deploy no cluster k3s da VPS. As imagens já vêm BUILDADAS do GHCR (job "build"
# do .github/workflows/deploy.yml roda num runner do GitHub) — aqui o containerd
# do k3s só as BAIXA. Buildar na VPS (1.9 GB) estourava a RAM (OOM), por isso o
# build saiu daqui. Idempotente — pode rodar quantas vezes for preciso.
#
# Requisitos na VPS: k3s, sudo sem senha p/ o k3s, e as variáveis
# SHA / OWNER / GHCR_USER / GHCR_TOKEN (injetadas pelo workflow de Deploy).
set -euo pipefail

cd "$(dirname "$0")/.."

: "${SHA:?SHA não definido (é injetado pelo workflow)}"
reg="ghcr.io/$(echo "${OWNER:-victor-teles-dev}" | tr '[:upper:]' '[:lower:]')"

echo ">> pull das imagens do GHCR direto no containerd do k3s"
# O containerd do k3s puxa do registro (caminho nativo do k8s): resolve só o
# manifesto linux/amd64 e ignora as attestations do buildkit. É diferente de
# `docker save | ctr import`, que quebrava com "content digest not found".
for img in backend-api url-shortener web-app; do
  src="$reg/ccl-${img}:$SHA"
  sudo k3s ctr images pull --platform linux/amd64 --user "${GHCR_USER:?}:${GHCR_TOKEN:?}" "$src"
  # Re-tag p/ o nome local que os manifests usam (ccl/<app>:dev, IfNotPresent).
  sudo k3s ctr images tag --force "$src" "docker.io/ccl/${img}:dev"
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
